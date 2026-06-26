package httpapi

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"remna-user-panel/internal/config"
	"remna-user-panel/internal/mail"
	appsettings "remna-user-panel/internal/settings"
)

const (
	emailCodePurposeVerify        = "verify"
	emailCodePurposeLogin         = "login"
	emailCodePurposePasswordReset = "password_reset"
	emailCodeLength               = 6
	emailCodeExpireMinutes        = 10
	emailCodeRateLimitSeconds     = 60
)

// Default email templates (Markdown). Admins can override via app_settings keys
// EMAIL_TEMPLATE_VERIFY, EMAIL_TEMPLATE_PASSWORD_RESET, EMAIL_TEMPLATE_LOGIN.
// To support multiple languages, configure these settings in the admin panel
// with translated text. i18n keys email_template_verify / email_template_password_reset /
// email_template_login are available in locale files as reference.
const defaultVerifyTemplate = `# {{.Brand}}

您的邮箱验证码如下

**{{.Code}}**

验证码 **{{.ExpireMinutes}}** 分钟内有效。如非本人操作请忽略。

---
这是来自 {{.Brand}} 的自动邮件。`

//nolint:gosec // Email template, not a credential.
const defaultPasswordResetTemplate = `# {{.Brand}}

您正在重置密码，验证码如下

**{{.Code}}**

验证码 **{{.ExpireMinutes}}** 分钟内有效。如非本人操作请忽略。

---
这是来自 {{.Brand}} 的自动邮件。`

const defaultLoginTemplate = `# {{.Brand}}

您的登录验证码如下

**{{.Code}}**

验证码 **{{.ExpireMinutes}}** 分钟内有效。如非本人操作请忽略。

---
这是来自 {{.Brand}} 的自动邮件。`

// English default templates used when the user's language is English and
// the admin has not configured custom templates.
const defaultVerifyTemplateEN = `# {{.Brand}}

Your verification code is:

**{{.Code}}**

This code expires in **{{.ExpireMinutes}}** minutes. If you did not request this, please ignore.

---
This is an automated email from {{.Brand}}.`

//nolint:gosec // Email template, not a credential.
const defaultPasswordResetTemplateEN = `# {{.Brand}}

You are resetting your password. Your verification code is:

**{{.Code}}**

This code expires in **{{.ExpireMinutes}}** minutes. If you did not request this, please ignore.

---
This is an automated email from {{.Brand}}.`

const defaultLoginTemplateEN = `# {{.Brand}}

Your login code is:

**{{.Code}}**

This code expires in **{{.ExpireMinutes}}** minutes. If you did not request this, please ignore.

---
This is an automated email from {{.Brand}}.`

// emailTemplate returns the rendered email body for the given purpose.
// It first tries the admin-configured Markdown template, falling back to
// a language-appropriate default (English for "en" users, Chinese otherwise).
func emailTemplate(ctx context.Context, store appsettings.Store, purpose, _ string, fallback string, vars map[string]any) string {
	var key string
	switch purpose {
	case emailCodePurposeVerify:
		key = "EMAIL_TEMPLATE_VERIFY"
	case emailCodePurposePasswordReset:
		key = "EMAIL_TEMPLATE_PASSWORD_RESET"
	case emailCodePurposeLogin:
		key = "EMAIL_TEMPLATE_LOGIN"
	}
	if key != "" {
		if tmpl := store.String(ctx, key, ""); strings.TrimSpace(tmpl) != "" {
			return mail.RenderTemplate(tmpl, vars)
		}
	}
	return mail.RenderTemplate(fallback, vars)
}

// emailFallbackTemplate selects the language-appropriate default template.
func emailFallbackTemplate(purpose, language string) string {
	isEN := strings.HasPrefix(strings.ToLower(strings.TrimSpace(language)), "en")
	switch purpose {
	case emailCodePurposeVerify:
		if isEN {
			return defaultVerifyTemplateEN
		}
		return defaultVerifyTemplate
	case emailCodePurposePasswordReset:
		if isEN {
			return defaultPasswordResetTemplateEN
		}
		return defaultPasswordResetTemplate
	case emailCodePurposeLogin:
		if isEN {
			return defaultLoginTemplateEN
		}
		return defaultLoginTemplate
	default:
		return defaultVerifyTemplate
	}
}

// emailSubject returns a localized email subject line.
func emailSubject(brand, purpose, language string) string {
	isEN := strings.HasPrefix(strings.ToLower(strings.TrimSpace(language)), "en")
	switch purpose {
	case emailCodePurposeVerify:
		if isEN {
			return fmt.Sprintf("%s — Verification Code", brand)
		}
		return fmt.Sprintf("%s — 邮箱验证码", brand)
	case emailCodePurposePasswordReset:
		if isEN {
			return fmt.Sprintf("%s — Password Reset Code", brand)
		}
		return fmt.Sprintf("%s — 密码重置验证码", brand)
	case emailCodePurposeLogin:
		if isEN {
			return fmt.Sprintf("%s — Login Code", brand)
		}
		return fmt.Sprintf("%s — 登录验证码", brand)
	default:
		return brand
	}
}

// isMailEnabled checks if SMTP is configured and enabled.
func isMailEnabled(ctx context.Context, store appsettings.Store) bool {
	if !store.Bool(ctx, "SMTP_ENABLED", false) {
		return false
	}
	return mailerFromStore(ctx, store).IsConfigured()
}

// emailRequestHandler sends a verification code to bind an email to a user account.
func emailRequestHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := requireSession(w, r, settings, pool, true)
		if !ok {
			return
		}
		var payload struct {
			Email string `json:"email"`
		}
		if err := decodeJSONBody(r, &payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_json"})
			return
		}
		email := strings.ToLower(strings.TrimSpace(payload.Email))
		if email == "" || !strings.Contains(email, "@") {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_email"})
			return
		}

		store := appsettings.NewStore(pool)
		if !isMailEnabled(r.Context(), store) {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": "email_delivery_not_configured"})
			return
		}
		mailer := mailerFromStore(r.Context(), store)

		// Check if email is already taken by another user
		var existingUserID int64
		err := pool.QueryRow(r.Context(),
			"SELECT user_id FROM users WHERE LOWER(email)=$1 AND user_id!=$2", email, session.User.UserID,
		).Scan(&existingUserID)
		if err == nil {
			writeJSON(w, http.StatusConflict, map[string]any{"ok": false, "error": "email_already_taken"})
			return
		}
		if err != pgx.ErrNoRows {
			slog.Error("email check failed", "error", err)
		}

		// Rate limit: check last code sent
		if lastSentAt, err := lastCodeSentAt(r.Context(), pool, email, emailCodePurposeVerify); err == nil {
			if time.Since(lastSentAt) < emailCodeRateLimitSeconds*time.Second {
				writeJSON(w, http.StatusTooManyRequests, map[string]any{"ok": false, "error": "rate_limit", "retry_after_seconds": emailCodeRateLimitSeconds})
				return
			}
		}

		code, err := generateEmailCode(emailCodeLength)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "code_generation_failed"})
			return
		}
		if err := storeEmailCode(r.Context(), pool, email, code, emailCodePurposeVerify, &session.User.UserID, emailCodeExpireMinutes); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "code_store_failed"})
			return
		}

		brand := store.String(r.Context(), "BRAND_NAME", "Remna")
		lang := session.User.LanguageCode
		subject := emailSubject(brand, emailCodePurposeVerify, lang)
		body := emailTemplate(r.Context(), store, emailCodePurposeVerify, lang, emailFallbackTemplate(emailCodePurposeVerify, lang), map[string]any{
			"Brand":         brand,
			"Code":          code,
			"ExpireMinutes": emailCodeExpireMinutes,
		})
		if err := mailer.Send(mail.Message{
			To:       []string{email},
			Subject:  subject,
			BodyHTML: body,
		}); err != nil {
			slog.Error("send verification email failed", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "email_send_failed"})
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	}
}

// emailVerifyHandler verifies the code sent to bind an email.
func emailVerifyHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := requireSession(w, r, settings, pool, true)
		if !ok {
			return
		}
		var payload struct {
			Email string `json:"email"`
			Code  string `json:"code"`
		}
		if err := decodeJSONBody(r, &payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_json"})
			return
		}
		email := strings.ToLower(strings.TrimSpace(payload.Email))
		code := strings.TrimSpace(payload.Code)
		if email == "" || code == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_params"})
			return
		}

		valid, err := verifyEmailCode(r.Context(), pool, email, code, emailCodePurposeVerify)
		if err != nil || !valid {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_code"})
			return
		}

		// Update user email
		now := time.Now()
		_, err = pool.Exec(r.Context(),
			"UPDATE users SET email=$1, email_verified_at=$2 WHERE user_id=$3",
			email, now, session.User.UserID)
		if err != nil {
			slog.Error("update user email failed", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "email_update_failed"})
			return
		}

		recordMessageLog(r.Context(), pool, messageLogEntry{
			UserID:    session.User.UserID,
			EventType: "account_email_linked",
			Content:   email,
		})
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "email": email, "verified": true})
	}
}

// passwordRequestHandler sends a password reset code to the user's registered email.
func passwordRequestHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := requireSession(w, r, settings, pool, true)
		if !ok {
			return
		}
		store := appsettings.NewStore(pool)
		if !isMailEnabled(r.Context(), store) {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": "email_delivery_not_configured"})
			return
		}
		mailer := mailerFromStore(r.Context(), store)

		// Get user's email
		user, err := loadWebappUser(r.Context(), pool, session.User.UserID, settings)
		if err != nil || user.Email == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "no_email_set"})
			return
		}

		// Rate limit
		if lastSentAt, err := lastCodeSentAt(r.Context(), pool, user.Email, emailCodePurposePasswordReset); err == nil {
			if time.Since(lastSentAt) < emailCodeRateLimitSeconds*time.Second {
				writeJSON(w, http.StatusTooManyRequests, map[string]any{"ok": false, "error": "rate_limit", "retry_after_seconds": emailCodeRateLimitSeconds})
				return
			}
		}

		code, err := generateEmailCode(emailCodeLength)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "code_generation_failed"})
			return
		}
		if err := storeEmailCode(r.Context(), pool, user.Email, code, emailCodePurposePasswordReset, &session.User.UserID, emailCodeExpireMinutes); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "code_store_failed"})
			return
		}

		brand := store.String(r.Context(), "BRAND_NAME", "Remna")
		lang := session.User.LanguageCode
		subject := emailSubject(brand, emailCodePurposePasswordReset, lang)
		body := emailTemplate(r.Context(), store, emailCodePurposePasswordReset, lang, emailFallbackTemplate(emailCodePurposePasswordReset, lang), map[string]any{
			"Brand":         brand,
			"Code":          code,
			"ExpireMinutes": emailCodeExpireMinutes,
		})
		if err := mailer.Send(mail.Message{
			To:       []string{user.Email},
			Subject:  subject,
			BodyHTML: body,
		}); err != nil {
			slog.Error("send password reset email failed", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "email_send_failed"})
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	}
}

// passwordConfirmHandler confirms password reset with code and new password.
func passwordConfirmHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := requireSession(w, r, settings, pool, true)
		if !ok {
			return
		}
		var payload struct {
			Code        string `json:"code"`
			NewPassword string `json:"new_password"`
		}
		if err := decodeJSONBody(r, &payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_json"})
			return
		}
		code := strings.TrimSpace(payload.Code)
		newPassword := strings.TrimSpace(payload.NewPassword)
		if code == "" || !isValidPassword(newPassword) {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_params", "password_requirements": "min_8_chars_upper_lower_digit"}) //nolint:gosec // G101: Error message, not a credential.
			return
		}

		user, err := loadWebappUser(r.Context(), pool, session.User.UserID, settings)
		if err != nil || user.Email == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "no_email_set"})
			return
		}

		valid, err := verifyEmailCode(r.Context(), pool, user.Email, code, emailCodePurposePasswordReset)
		if err != nil || !valid {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_code"})
			return
		}

		// Hash and store the new password
		passwordHash, err := hashPassword(newPassword)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "password_hash_failed"})
			return
		}
		now := time.Now()
		_, err = pool.Exec(r.Context(),
			"UPDATE users SET password_hash=$1, password_set_at=$2 WHERE user_id=$3",
			passwordHash, now, session.User.UserID)
		if err != nil {
			slog.Error("update password failed", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "password_update_failed"})
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	}
}

// emailLoginRequestHandler sends a login code to an email (public endpoint, no session required).
func emailLoginRequestHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			Email        string `json:"email"`
			Language     string `json:"language"`
			ReferralCode string `json:"referral_code"`
		}
		if err := decodeJSONBody(r, &payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_json"})
			return
		}
		email := strings.ToLower(strings.TrimSpace(payload.Email))
		if email == "" || !strings.Contains(email, "@") {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_email"})
			return
		}

		store := appsettings.NewStore(pool)
		if !isMailEnabled(r.Context(), store) {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": "email_delivery_not_configured"})
			return
		}
		mailer := mailerFromStore(r.Context(), store)

		// Check if user exists with this email
		var userID int64
		err := pool.QueryRow(r.Context(),
			"SELECT user_id FROM users WHERE LOWER(email)=$1", email,
		).Scan(&userID)
		if err != nil {
			if err == pgx.ErrNoRows {
				// Don't reveal whether email exists; silently succeed
				writeJSON(w, http.StatusOK, map[string]any{"ok": true})
				return
			}
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "internal_error"})
			return
		}

		// Rate limit
		if lastSentAt, err := lastCodeSentAt(r.Context(), pool, email, emailCodePurposeLogin); err == nil {
			if time.Since(lastSentAt) < emailCodeRateLimitSeconds*time.Second {
				writeJSON(w, http.StatusOK, map[string]any{"ok": true})
				return
			}
		}

		code, err := generateEmailCode(emailCodeLength)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "code_generation_failed"})
			return
		}
		if err := storeEmailCode(r.Context(), pool, email, code, emailCodePurposeLogin, &userID, emailCodeExpireMinutes); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "code_store_failed"})
			return
		}

		brand := store.String(r.Context(), "BRAND_NAME", "Remna")
		lang := normalizeWebLanguage(payload.Language, effectiveDefaultLanguage(r.Context(), pool, settings))
		if lang == "" && userID != 0 {
			_ = pool.QueryRow(r.Context(), "SELECT COALESCE(language_code,'') FROM users WHERE user_id=$1", userID).Scan(&lang)
		}
		subject := emailSubject(brand, emailCodePurposeLogin, lang)
		body := emailTemplate(r.Context(), store, emailCodePurposeLogin, lang, emailFallbackTemplate(emailCodePurposeLogin, lang), map[string]any{
			"Brand":         brand,
			"Code":          code,
			"ExpireMinutes": emailCodeExpireMinutes,
		})
		if err := mailer.Send(mail.Message{
			To:       []string{email},
			Subject:  subject,
			BodyHTML: body,
		}); err != nil {
			slog.Error("send login email failed", "error", err)
		}

		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	}
}

// emailLoginVerifyHandler verifies the login code and creates a session (public endpoint).
func emailLoginVerifyHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			Email        string `json:"email"`
			Code         string `json:"code"`
			Language     string `json:"language"`
			ReferralCode string `json:"referral_code"`
		}
		if err := decodeJSONBody(r, &payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_json"})
			return
		}
		email := strings.ToLower(strings.TrimSpace(payload.Email))
		code := strings.TrimSpace(payload.Code)
		if email == "" || code == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_params"})
			return
		}

		valid, err := verifyEmailCode(r.Context(), pool, email, code, emailCodePurposeLogin)
		if err != nil || !valid {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_code"})
			return
		}

		// Look up user
		var userID int64
		err = pool.QueryRow(r.Context(),
			"SELECT user_id FROM users WHERE LOWER(email)=$1", email,
		).Scan(&userID)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "user_not_found"})
			return
		}
		bindReferralCode(r.Context(), pool, userID, payload.ReferralCode)

		// Persist language preference if provided.
		if lang := normalizeWebLanguage(payload.Language, effectiveDefaultLanguage(r.Context(), pool, settings)); lang != "" {
			_, _ = pool.Exec(r.Context(), "UPDATE users SET language_code=$1 WHERE user_id=$2 AND language_code<>$1", lang, userID)
		}

		// Create session
		manager := webappSessionManager(settings)
		token, csrf, err := manager.Sign(userID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "session_create_failed"})
			return
		}

		setSessionCookies(w, r, token, csrf)
		recordMessageLog(r.Context(), pool, messageLogEntry{
			UserID:    userID,
			EventType: "email_code_login",
			Content:   email,
		})
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "csrf_token": csrf})
	}
}

// emailMagicLinkHandler validates a magic link token and creates a session.
func emailMagicLinkHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			Email        string `json:"email"`
			Token        string `json:"token"`
			ReferralCode string `json:"referral_code"`
		}
		if err := decodeJSONBody(r, &payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_json"})
			return
		}
		email := strings.ToLower(strings.TrimSpace(payload.Email))
		token := strings.TrimSpace(payload.Token)
		if email == "" || token == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_params"})
			return
		}

		store := appsettings.NewStore(pool)
		if !isMailEnabled(r.Context(), store) {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": "email_delivery_not_configured"})
			return
		}

		// Verify magic link token
		valid, err := verifyEmailCode(r.Context(), pool, email, token, emailCodePurposeLogin)
		if err != nil || !valid {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_token"})
			return
		}

		// Look up user
		var userID int64
		err = pool.QueryRow(r.Context(),
			"SELECT user_id FROM users WHERE LOWER(email)=$1", email,
		).Scan(&userID)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "user_not_found"})
			return
		}
		bindReferralCode(r.Context(), pool, userID, payload.ReferralCode)

		// Create session
		manager := webappSessionManager(settings)
		sessionToken, csrf, err := manager.Sign(userID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "session_create_failed"})
			return
		}

		setSessionCookies(w, r, sessionToken, csrf)
		recordMessageLog(r.Context(), pool, messageLogEntry{
			UserID:    userID,
			EventType: "email_magic_login",
			Content:   email,
		})
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "csrf_token": csrf})
	}
}

// passwordLoginHandler authenticates a user (admin or regular) via email and password.
func passwordLoginHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	const (
		passwordLoginMaxAttempts = 5
		passwordLoginLockMinutes = 15
	)

	return func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := decodeJSONBody(r, &payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_json"})
			return
		}
		email := strings.ToLower(strings.TrimSpace(payload.Email))
		password := payload.Password
		if email == "" || !strings.Contains(email, "@") || password == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_params"})
			return
		}

		// Rate limiting: count failed attempts per email since the lock window.
		var failedAttempts int
		_ = pool.QueryRow(r.Context(),
			`SELECT COUNT(*) FROM email_verification_codes
			 WHERE email=$1 AND purpose='password_login_failed'
			 AND created_at > NOW() - ($2 || ' minutes')::INTERVAL`,
			email, strconv.Itoa(passwordLoginLockMinutes),
		).Scan(&failedAttempts)
		if failedAttempts >= passwordLoginMaxAttempts {
			writeJSON(w, http.StatusTooManyRequests, map[string]any{
				"ok":                  false,
				"error":               "too_many_attempts",
				"retry_after_minutes": passwordLoginLockMinutes,
			})
			return
		}

		// Look up user by email and retrieve password hash and ban status.
		var userID int64
		var passwordHash sql.NullString
		var isBanned bool
		err := pool.QueryRow(r.Context(),
			"SELECT user_id, password_hash, is_banned FROM users WHERE LOWER(email)=$1", email,
		).Scan(&userID, &passwordHash, &isBanned)
		if err != nil {
			if err == pgx.ErrNoRows {
				// Admin bootstrap: 如果用户不存在但邮箱密码匹配 ADMIN_EMAIL / ADMIN_PASSWORD，
				// 则自动创建管理员用户。
				adminEmail := strings.ToLower(strings.TrimSpace(settings.AdminEmail))
				adminPassword := settings.AdminPassword
				if email == adminEmail && password == adminPassword {
					passwordHashStr, hashErr := hashPassword(password)
					if hashErr != nil {
						writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "internal_error"})
						return
					}
					uid := adminEmailHash(email)
					err = pool.QueryRow(r.Context(), `
INSERT INTO users (user_id, telegram_id, email, password_hash, password_set_at, language_code, registration_date)
VALUES ($1, NULL, $2, $3, NOW(), 'zh', NOW())
ON CONFLICT (email) DO UPDATE SET password_hash=EXCLUDED.password_hash, password_set_at=NOW()
RETURNING user_id`, uid, email, passwordHashStr).Scan(&userID)
					if err != nil {
						slog.Error("admin user auto-create failed", "error", err)
						writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "internal_error"})
						return
					}
					slog.Info("admin user auto-created from env vars", "email", email, "user_id", userID)
					// 用户已创建/更新，继续执行后续登录流程。
					passwordHash = sql.NullString{String: passwordHashStr, Valid: true}
					isBanned = false
				} else {
					// Record failed attempt even when user doesn't exist (prevents enumeration).
					_ = storeEmailCode(r.Context(), pool, email, "failed", "password_login_failed", nil, passwordLoginLockMinutes)
					writeJSON(w, http.StatusUnauthorized, map[string]any{"ok": false, "error": "password_login_failed", "fallback": "email_code"})
					return
				}
			} else {
				slog.Error("password login lookup failed", "error", err)
				writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "internal_error"})
				return
			}
		}

		if isBanned {
			writeJSON(w, http.StatusForbidden, map[string]any{"ok": false, "error": "banned"})
			return
		}

		passwordValid, needsRehash := VerifyPassword(password, passwordHash.String)
		if !passwordHash.Valid || !passwordValid {
			_ = storeEmailCode(r.Context(), pool, email, "failed", "password_login_failed", &userID, passwordLoginLockMinutes)
			writeJSON(w, http.StatusUnauthorized, map[string]any{"ok": false, "error": "password_login_failed", "fallback": "email_code"})
			return
		}

		// Auto-migrate legacy SHA-256 hashes to bcrypt.
		if needsRehash {
			newHash, hashErr := hashPassword(password)
			if hashErr == nil {
				_, _ = pool.Exec(r.Context(),
					"UPDATE users SET password_hash=$1 WHERE user_id=$2", newHash, userID)
			}
		}

		// Create session
		manager := webappSessionManager(settings)
		token, csrf, err := manager.Sign(userID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "session_create_failed"})
			return
		}

		setSessionCookies(w, r, token, csrf)
		recordMessageLog(r.Context(), pool, messageLogEntry{
			UserID:    userID,
			EventType: "email_password_login",
			Content:   email,
		})
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "csrf_token": csrf})
	}
}

// mailerFromStore reads SMTP configuration from app_settings and builds a Mailer.
func mailerFromStore(ctx context.Context, store appsettings.Store) *mail.Mailer {
	return mail.NewMailer(mailerConfigFromSettings(ctx, store))
}

// generateEmailCode produces a numeric verification code of the given length.
func generateEmailCode(length int) (string, error) {
	code := make([]byte, length)
	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		code[i] = byte('0' + n.Int64()) //nolint:gosec // G115: n.Int64() is 0-9 from rand.Int(10).
	}
	return string(code), nil
}

// storeEmailCode persists a verification code to the database.
func storeEmailCode(ctx context.Context, pool *pgxpool.Pool, email, code, purpose string, userID *int64, expireMinutes int) error {
	expiresAt := time.Now().Add(time.Duration(expireMinutes) * time.Minute)
	_, err := pool.Exec(ctx, `
INSERT INTO email_verification_codes (email, code, purpose, user_id, expires_at)
VALUES ($1, $2, $3, $4, $5)`,
		email, code, purpose, userID, expiresAt)
	return err
}

// verifyEmailCode checks if a code is valid and marks it as used.
func verifyEmailCode(ctx context.Context, pool *pgxpool.Pool, email, code, purpose string) (bool, error) {
	var id int64
	err := pool.QueryRow(ctx, `
DELETE FROM email_verification_codes
WHERE id = (
	SELECT id FROM email_verification_codes
	WHERE email=$1 AND code=$2 AND purpose=$3 AND used=FALSE AND expires_at > NOW()
	ORDER BY created_at DESC LIMIT 1
)
RETURNING id`, email, code, purpose).Scan(&id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return id > 0, nil
}

// lastCodeSentAt returns the creation time of the most recent code for the given email and purpose.
func lastCodeSentAt(ctx context.Context, pool *pgxpool.Pool, email, purpose string) (time.Time, error) {
	var ts time.Time
	err := pool.QueryRow(ctx, `
SELECT created_at FROM email_verification_codes
WHERE email=$1 AND purpose=$2
ORDER BY created_at DESC LIMIT 1`, email, purpose).Scan(&ts)
	return ts, err
}

// hashPassword creates a bcrypt hash of the password (cost=12).
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// passwordNeedsRehash returns true when a stored hash uses the legacy SHA-256 format
// and should be upgraded to bcrypt.
func passwordNeedsRehash(storedHash string) bool {
	return strings.HasPrefix(storedHash, "sha256:")
}

// VerifyPassword checks if the given password matches the stored hash.
// Supports both legacy sha256: prefixed hashes (auto-migrated to bcrypt on
// first successful verification) and bcrypt hashes.
//
// Returns (valid, needsRehash).  When valid is true and needsRehash is true,
// the caller MUST rehash the password with bcrypt and persist the new hash.
func VerifyPassword(password, storedHash string) (valid bool, needsRehash bool) {
	if password == "" || storedHash == "" {
		return false, false
	}
	if strings.HasPrefix(storedHash, "sha256:") {
		// Legacy SHA-256 format: "sha256:<hex>".  SHA-256 is not a
		// suitable password hash (no salt, too fast).  This code path
		// exists SOLELY for backward compatibility during migration;
		// all legacy hashes are upgraded to bcrypt (cost=12) on
		// successful authentication.
		if verifyLegacySHA256(password, storedHash) {
			return true, true // valid, needs upgrade
		}
		return false, false
	}
	// Bcrypt format: "$2a$..."
	return bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password)) == nil, false
}

// verifyLegacySHA256 compares a plaintext password against a legacy
// "sha256:<hex>" hash using constant-time comparison.  This function
// exists ONLY for backward compatibility during the SHA-256 → bcrypt
// migration and MUST NOT be used for new password hashes.
func verifyLegacySHA256(password, storedHash string) bool {
	h := sha256.Sum256([]byte(password))
	expected := strings.TrimPrefix(storedHash, "sha256:")
	return subtle.ConstantTimeCompare([]byte(hex.EncodeToString(h[:])), []byte(expected)) == 1
}

// isValidPassword validates password strength: minimum 8 characters,
// at least one uppercase letter, one lowercase letter, and one digit.
func isValidPassword(password string) bool {
	if len(password) < 8 || len(password) > 128 {
		return false
	}
	hasUpper := false
	hasLower := false
	hasDigit := false
	for _, r := range password {
		switch {
		case r >= 'A' && r <= 'Z':
			hasUpper = true
		case r >= 'a' && r <= 'z':
			hasLower = true
		case r >= '0' && r <= '9':
			hasDigit = true
		}
	}
	return hasUpper && hasLower && hasDigit
}

// adminEmailHash 将 ADMIN_EMAIL 映射为一个稳定的 user_id（非负 int64），
// 用于首次启动时自动创建管理员用户。
func adminEmailHash(email string) int64 {
	h := sha256.Sum256([]byte(strings.TrimSpace(strings.ToLower(email))))
	// 取前 8 字节并保证非负
	var id int64
	for i := 0; i < 8; i++ {
		id = (id << 8) | int64(h[i])
	}
	if id < 0 {
		id = -id
	}
	if id == 0 {
		id = 1
	}
	// 以管理员邮箱哈希的高位确保不会与 Telegram ID 冲突（Telegram ID 为 10 位以内正整数）。
	// 加一个偏移量使 ID 落在 2^63 - 1 范围内但足够高，避免碰撞。
	const emailIDOffset = 1 << 56
	return (id % (1<<56 - 1)) + emailIDOffset
}
