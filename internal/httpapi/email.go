package httpapi

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"remna-user-panel/internal/config"
	"remna-user-panel/internal/mail"
	appsettings "remna-user-panel/internal/settings"
)

const (
	emailCodePurposeVerify       = "verify"
	emailCodePurposeLogin        = "login"
	emailCodePurposePasswordReset = "password_reset"
	emailCodeLength              = 6
	emailCodeExpireMinutes       = 10
	emailCodeRateLimitSeconds    = 60
)

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
		subject := fmt.Sprintf("%s — 邮箱验证码", brand)
		body := buildVerificationEmailHTML(brand, code, emailCodeExpireMinutes)
		if err := mailer.Send(mail.Message{
			To:      []string{email},
			Subject: subject,
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
		subject := fmt.Sprintf("%s — 密码重置验证码", brand)
		body := buildPasswordResetEmailHTML(brand, code, emailCodeExpireMinutes)
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
		if code == "" || len(newPassword) < 6 {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_params"})
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
		subject := fmt.Sprintf("%s — 登录验证码", brand)
		body := buildLoginEmailHTML(brand, code, emailCodeExpireMinutes)
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

		// Create session
		manager := webappSessionManager(settings)
		token, csrf, err := manager.Sign(userID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "session_create_failed"})
			return
		}

		setSessionCookies(w, r, token, csrf)
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	}
}

// emailMagicLinkHandler validates a magic link token and creates a session.
func emailMagicLinkHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			Email string `json:"email"`
			Token string `json:"token"`
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
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
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
		code[i] = byte('0' + n.Int64())
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

// buildVerificationEmailHTML generates the HTML body for an email verification code.
func buildVerificationEmailHTML(brand, code string, expireMinutes int) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family: Arial, sans-serif; max-width: 480px; margin: 0 auto; padding: 20px;">
	<div style="background: #f8f9fa; border-radius: 12px; padding: 30px; text-align: center;">
		<h2 style="color: #1a1a2e; margin: 0 0 10px;">%s</h2>
		<p style="color: #666; font-size: 15px; margin: 0 0 24px;">您的邮箱验证码如下</p>
		<div style="background: #fff; border-radius: 8px; padding: 16px 24px; margin: 0 auto 24px; display: inline-block;">
			<span style="font-size: 32px; font-weight: 700; letter-spacing: 6px; color: #2563eb;">%s</span>
		</div>
		<p style="color: #999; font-size: 13px; margin: 0;">验证码 %d 分钟内有效。如非本人操作请忽略。</p>
	</div>
	<p style="color: #aaa; font-size: 12px; text-align: center; margin: 16px 0 0;">这是来自 %s 的自动邮件。</p>
</body>
</html>`, brand, code, expireMinutes, brand)
}

// buildPasswordResetEmailHTML generates the HTML body for a password reset code.
func buildPasswordResetEmailHTML(brand, code string, expireMinutes int) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family: Arial, sans-serif; max-width: 480px; margin: 0 auto; padding: 20px;">
	<div style="background: #f8f9fa; border-radius: 12px; padding: 30px; text-align: center;">
		<h2 style="color: #1a1a2e; margin: 0 0 10px;">%s</h2>
		<p style="color: #666; font-size: 15px; margin: 0 0 24px;">您正在重置密码，验证码如下</p>
		<div style="background: #fff; border-radius: 8px; padding: 16px 24px; margin: 0 auto 24px; display: inline-block;">
			<span style="font-size: 32px; font-weight: 700; letter-spacing: 6px; color: #dc2626;">%s</span>
		</div>
		<p style="color: #999; font-size: 13px; margin: 0;">验证码 %d 分钟内有效。如非本人操作请忽略。</p>
	</div>
	<p style="color: #aaa; font-size: 12px; text-align: center; margin: 16px 0 0;">这是来自 %s 的自动邮件。</p>
</body>
</html>`, brand, code, expireMinutes, brand)
}

// buildLoginEmailHTML generates the HTML body for a login code.
func buildLoginEmailHTML(brand, code string, expireMinutes int) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family: Arial, sans-serif; max-width: 480px; margin: 0 auto; padding: 20px;">
	<div style="background: #f8f9fa; border-radius: 12px; padding: 30px; text-align: center;">
		<h2 style="color: #1a1a2e; margin: 0 0 10px;">%s</h2>
		<p style="color: #666; font-size: 15px; margin: 0 0 24px;">您的登录验证码如下</p>
		<div style="background: #fff; border-radius: 8px; padding: 16px 24px; margin: 0 auto 24px; display: inline-block;">
			<span style="font-size: 32px; font-weight: 700; letter-spacing: 6px; color: #2563eb;">%s</span>
		</div>
		<p style="color: #999; font-size: 13px; margin: 0;">验证码 %d 分钟内有效。如非本人操作请忽略。</p>
	</div>
	<p style="color: #aaa; font-size: 12px; text-align: center; margin: 16px 0 0;">这是来自 %s 的自动邮件。</p>
</body>
</html>`, brand, code, expireMinutes, brand)
}

// hashPassword creates a SHA-256 hash of the password.
func hashPassword(password string) (string, error) {
	return fmt.Sprintf("sha256:%s", sha256Hex(password)), nil
}

// sha256Hex returns the hex-encoded SHA-256 hash of the input.
func sha256Hex(input string) string {
	h := sha256.Sum256([]byte(input))
	return hex.EncodeToString(h[:])
}

// VerifyPassword checks if the given password matches the stored hash.
func VerifyPassword(password, storedHash string) bool {
	if !strings.HasPrefix(storedHash, "sha256:") {
		return false
	}
	return sha256Hex(password) == strings.TrimPrefix(storedHash, "sha256:")
}
