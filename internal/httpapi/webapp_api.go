package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"remna-user-panel/internal/auth"
	"remna-user-panel/internal/config"
	"remna-user-panel/internal/fx"
	"remna-user-panel/internal/mail"
	"remna-user-panel/internal/payments"
	"remna-user-panel/internal/remnawave"
	appsettings "remna-user-panel/internal/settings"
	"remna-user-panel/internal/tariffs"
)

const (
	sessionCookieName       = "rw_webapp_session"
	csrfCookieName          = "rw_webapp_csrf"
	telegramNonceCookieName = "rw_telegram_login_nonce"
)

type sessionContext struct {
	Claims auth.SessionClaims
	User   webappUser
}

type webappUser struct {
	UserID        int64  `json:"user_id"`
	TelegramID    int64  `json:"telegram_id,omitempty"`
	Username      string `json:"username,omitempty"`
	Email         string `json:"email,omitempty"`
	FirstName     string `json:"first_name,omitempty"`
	LastName      string `json:"last_name,omitempty"`
	LanguageCode  string `json:"language_code"`
	PhotoURL      string `json:"telegram_photo_url,omitempty"`
	PanelUserUUID string `json:"panel_user_uuid,omitempty"`
	IsAdmin       bool   `json:"is_admin"`
}

type telegramAuthPayload struct {
	InitData     string `json:"init_data"`
	IDToken      string `json:"id_token"`
	Nonce        string `json:"nonce"`
	ReferralCode string `json:"referral_code"`
}

func validateTelegramAuthPayload(ctx context.Context, r *http.Request, settings config.Settings, pool *pgxpool.Pool, payload telegramAuthPayload) (auth.TelegramUser, error) {
	if strings.TrimSpace(payload.InitData) != "" {
		return auth.ValidateTelegramInitData(payload.InitData, settings.BotToken, 24*time.Hour)
	}
	if strings.TrimSpace(payload.IDToken) == "" {
		return auth.TelegramUser{}, errors.New("missing_telegram_auth")
	}
	cookie, cookieErr := r.Cookie(telegramNonceCookieName)
	clientID := appsettings.NewStore(pool).String(ctx, "TELEGRAM_LOGIN_CLIENT_ID", os.Getenv("TELEGRAM_LOGIN_CLIENT_ID"))
	if cookieErr != nil || cookie.Value == "" || payload.Nonce == "" || cookie.Value != payload.Nonce || clientID == "" {
		return auth.TelegramUser{}, errors.New("invalid_telegram_nonce")
	}
	return auth.ValidateTelegramIDToken(ctx, payload.IDToken, clientID, payload.Nonce)
}

func authTokenHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if pool == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": "database_not_configured"})
			return
		}
		var payload telegramAuthPayload
		if err := decodeJSONBody(r, &payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_json"})
			return
		}
		tgUser, err := validateTelegramAuthPayload(r.Context(), r, settings, pool, payload)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"ok": false, "error": "invalid_telegram_init_data"})
			return
		}
		language := normalizeWebLanguage(tgUser.LanguageCode, effectiveDefaultLanguage(r.Context(), pool, settings))
		if err := upsertTelegramUser(r.Context(), pool, tgUser, language); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "user_upsert_failed"})
			return
		}
		bindReferralCode(r.Context(), pool, tgUser.ID, payload.ReferralCode)
		manager := webappSessionManager(settings)
		token, csrf, err := manager.Sign(tgUser.ID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "session_sign_failed"})
			return
		}
		setSessionCookies(w, r, token, csrf)
		http.SetCookie(w, &http.Cookie{Name: telegramNonceCookieName, Value: "", Path: "/", MaxAge: -1, HttpOnly: true, Secure: requestIsHTTPS(r), SameSite: http.SameSiteLaxMode})
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":         true,
			"csrf_token": csrf,
			"user": webappUser{
				UserID:       tgUser.ID,
				TelegramID:   tgUser.ID,
				Username:     tgUser.Username,
				FirstName:    tgUser.FirstName,
				LastName:     tgUser.LastName,
				LanguageCode: language,
				PhotoURL:     tgUser.PhotoURL,
				IsAdmin:      isAdminID(settings.AdminIDs, tgUser.ID),
			},
		})
	}
}

func telegramLoginNonceHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientID := appsettings.NewStore(pool).String(r.Context(), "TELEGRAM_LOGIN_CLIENT_ID", os.Getenv("TELEGRAM_LOGIN_CLIENT_ID"))
		if clientID == "" {
			writeJSON(w, http.StatusConflict, map[string]any{"ok": false, "error": "telegram_login_not_configured"})
			return
		}
		raw := make([]byte, 32)
		if _, err := rand.Read(raw); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "nonce_failed"})
			return
		}
		nonce := base64.RawURLEncoding.EncodeToString(raw)
		http.SetCookie(w, &http.Cookie{Name: telegramNonceCookieName, Value: nonce, Path: "/", MaxAge: 300, HttpOnly: true, Secure: requestIsHTTPS(r), SameSite: http.SameSiteLaxMode})
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "nonce": nonce, "client_id": clientID})
	}
}

func logoutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clearCookie(w, r, sessionCookieName)
		clearCookie(w, r, csrfCookieName)
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	}
}

func meHandler(settings config.Settings, pool *pgxpool.Pool, registry *payments.Registry, panel *remnawave.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := requireSession(w, r, settings, pool, false)
		if !ok {
			return
		}
		catalog, err := tariffs.Load("data/tariffs.json")
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "tariffs_load_failed"})
			return
		}
		rate := fx.NewService(appsettings.NewStore(pool)).USDCNY(r.Context())
		store := appsettings.NewStore(pool)
		plans := tariffs.WithCNYDisplay(
			catalog.Plans(session.User.LanguageCode, effectiveDefaultCurrency(r.Context(), settings, pool)),
			rate.Rate,
			rate.Source,
			rate.UpdatedAt,
		)
		plans = tariffs.WithStarsPrice(plans, store.Float(r.Context(), "STARS_USD_RATE", settings.StarsUSDRate))
		methods := []payments.Method{}
		if registry != nil {
			methods = registry.Methods(r.Context(), session.User.LanguageCode, session.User.IsAdmin)
		}
		subscription := map[string]any{"active": false}
		if panel != nil && panel.Configured(r.Context()) {
			panelUser, found, err := panelUserForWebUser(r.Context(), pool, panel, session.User)
			if err != nil {
				slog.Warn("failed to fetch panel user for /api/me, continuing without subscription data", "error", err, "user_id", session.User.UserID)
			} else if found {
				subscription = subscriptionFromPanelUser(r.Context(), pool, session.User, panelUser)
			}
		}
		if active, _ := subscription["active"].(bool); active {
			if token, tokenErr := ensureSubscriptionShareToken(r.Context(), pool, session.User.UserID); tokenErr == nil && token != "" {
				shareURL := strings.TrimRight(webappPublicBaseURL(r, settings), "/") + "/s/" + token
				subscription["install_share_token"] = token
				subscription["install_share_url"] = shareURL
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":                      true,
			"user":                    session.User,
			"subscription":            subscription,
			"settings":                webappFeatureSettings(r.Context(), settings, pool, panel, session.User),
			"referral":                referralPayload(r.Context(), settings, pool, session.User),
			"plans":                   plans,
			"payment_methods":         methods,
			"notification_prefs":      loadUserNotificationPrefs(r.Context(), pool, session.User.UserID),
		})
	}
}

func ensureSubscriptionShareToken(ctx context.Context, pool *pgxpool.Pool, userID int64) (string, error) {
	if pool == nil || userID == 0 {
		return "", errors.New("database_not_configured")
	}
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	token := fmt.Sprintf("%x", raw)
	var stored string
	err := pool.QueryRow(ctx, `INSERT INTO subscription_share_tokens(token,user_id) VALUES($1,$2)
ON CONFLICT(user_id) DO UPDATE SET user_id=EXCLUDED.user_id RETURNING token`, token, userID).Scan(&stored)
	return stored, err
}

func webappPublicBaseURL(r *http.Request, settings config.Settings) string {
	if configured := strings.TrimSpace(settings.SubscriptionMiniApp); configured != "" {
		return configured
	}
	scheme := "http"
	if requestIsHTTPS(r) {
		scheme = "https"
	}
	host := strings.TrimSpace(r.Host)
	return scheme + "://" + host
}

func createPaymentHandler(settings config.Settings, pool *pgxpool.Pool, registry *payments.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := requireSession(w, r, settings, pool, true)
		if !ok {
			return
		}
		if registry == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": "payments_not_configured"})
			return
		}
		var payload struct {
			Method   string `json:"method"`
			PlanHash string `json:"plan_hash"`
		}
		if err := decodeJSONBody(r, &payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_json"})
			return
		}
		if strings.TrimSpace(payload.PlanHash) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "plan_hash_required"})
			return
		}
		catalog, err := tariffs.Load("data/tariffs.json")
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "tariffs_load_failed"})
			return
		}
		defaultCurrency := effectiveDefaultCurrency(r.Context(), settings, pool)
		if len(catalog.Plans(session.User.LanguageCode, defaultCurrency)) == 0 {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "tariffs_not_configured"})
			return
		}
		plan, found := catalog.FindPlanByHash(payload.PlanHash, session.User.LanguageCode, defaultCurrency)
		if !found {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "plan_not_found"})
			return
		}
		rate := fx.NewService(appsettings.NewStore(pool)).USDCNY(r.Context())
		store := appsettings.NewStore(pool)
		plan = tariffs.WithCNYDisplay([]tariffs.Plan{plan}, rate.Rate, rate.Source, rate.UpdatedAt)[0]
		plan = tariffs.WithStarsPrice([]tariffs.Plan{plan}, store.Float(r.Context(), "STARS_USD_RATE", settings.StarsUSDRate))[0]
		providerAmount, providerCurrency, err := providerCheckoutAmount(payload.Method, plan, rate)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "payment_method_not_supported"})
			return
		}
		planSnapshot, err := json.Marshal(plan)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "plan_snapshot_failed"})
			return
		}
		description := plan.Title
		if plan.Months > 0 {
			description = fmt.Sprintf("%s %d month", plan.Title, plan.Months)
		} else if plan.TrafficGB > 0 {
			description = fmt.Sprintf("%s %s GB", plan.Title, compactFloat(plan.TrafficGB))
		}
		response, err := registry.Create(r.Context(), payments.CreateOrderRequest{
			UserID:           session.User.UserID,
			MethodID:         payload.Method,
			Amount:           providerAmount,
			Currency:         providerCurrency,
			BaseAmount:       plan.BaseAmount,
			BaseCurrency:     plan.BaseCurrency,
			DisplayCNYAmount: plan.DisplayCNYAmount,
			FXRate:           plan.FXRate,
			FXSource:         plan.FXSource,
			FXUpdatedAt:      rate.UpdatedAt,
			PlanHash:         plan.PlanHash,
			PlanSnapshot:     planSnapshot,
			Description:      description,
			TariffKey:        plan.TariffKey,
			SaleMode:         plan.SaleMode,
			Months:           plan.Months,
			TrafficGB:        plan.TrafficGB,
			ClientIP:         clientIP(r),
			Language:         session.User.LanguageCode,
		})
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": paymentErrorCode(err)})
			return
		}
		writeJSON(w, http.StatusOK, response)
	}
}

func paymentStatusHandler(settings config.Settings, pool *pgxpool.Pool, registry *payments.Registry, panel *remnawave.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := requireSession(w, r, settings, pool, false)
		if !ok {
			return
		}
		paymentID, err := parsePositiveInt64(chi.URLParam(r, "payment_id"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_payment_id"})
			return
		}
		order, err := registry.GetForUser(r.Context(), session.User.UserID, paymentID)
		if err != nil {
			status := http.StatusInternalServerError
			code := "payment_load_failed"
			if errors.Is(err, pgx.ErrNoRows) {
				status = http.StatusNotFound
				code = "payment_not_found"
			}
			writeJSON(w, status, map[string]any{"ok": false, "error": code})
			return
		}
		if paymentOrderPaid(order) {
			_ = provisionPaidOrder(r.Context(), settings, pool, panel, order)
			if refreshed, err := registry.GetForUser(r.Context(), session.User.UserID, paymentID); err == nil {
				order = refreshed
			}
		}
		writeJSON(w, http.StatusOK, paymentStatusPayload(order))
	}
}

func adminSettingsHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	store := appsettings.NewStore(pool)
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, r.Method != http.MethodGet); !ok {
			return
		}
		if r.Method == http.MethodGet {
			sections, err := adminSettingsSections(r.Context(), settings, store)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "settings_catalog_invalid"})
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{
				"ok":       true,
				"features": []string{"general", "remnawave", "features", "notifications", "telemetry", "payments", "mail", "subscription_guides", "appearance"},
				"sections": sections,
			})
			return
		}
		var payload struct {
			Updates map[string]any `json:"updates"`
			Deletes []string       `json:"deletes"`
		}
		if err := decodeJSONBody(r, &payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_json"})
			return
		}
		sections, catalogErr := adminSettingsSections(r.Context(), settings, store)
		if catalogErr != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "settings_catalog_invalid"})
			return
		}
		allowed := allowedAdminSettingKeys(sections)
		errorsByKey := map[string]string{}
		normalizedUpdates := map[string]any{}
		for key, value := range payload.Updates {
			if !allowed[key] {
				errorsByKey[key] = "unsupported_setting"
				continue
			}
			if _, locked := os.LookupEnv(key); locked {
				errorsByKey[key] = "setting_locked_by_env"
				continue
			}
			normalized, err := normalizeSettingValue(key, value)
			if err != nil {
				errorsByKey[key] = err.Error()
				continue
			}
			normalizedUpdates[key] = normalized
		}
		for _, key := range payload.Deletes {
			if !allowed[key] {
				errorsByKey[key] = "unsupported_setting"
				continue
			}
			if _, locked := os.LookupEnv(key); locked {
				errorsByKey[key] = "setting_locked_by_env"
				continue
			}
		}
		if len(errorsByKey) > 0 {
			status := http.StatusBadRequest
			for _, code := range errorsByKey {
				if code == "setting_locked_by_env" {
					status = http.StatusConflict
					break
				}
			}
			writeJSON(w, status, map[string]any{"ok": false, "errors": errorsByKey})
			return
		}
		for key, value := range normalizedUpdates {
			if err := store.Upsert(r.Context(), key, value); err != nil {
				errorsByKey[key] = "save_failed"
			}
		}
		for _, key := range payload.Deletes {
			if err := store.Delete(r.Context(), key); err != nil {
				errorsByKey[key] = "delete_failed"
			}
		}
		if len(errorsByKey) > 0 {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "errors": errorsByKey})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	}
}

func adminPaymentsListHandler(settings config.Settings, pool *pgxpool.Pool, registry *payments.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, false); !ok {
			return
		}
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
		orders, total, err := registry.List(r.Context(), page, pageSize)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "payments_load_failed"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "payments": orders, "total": total})
	}
}

func adminPaymentDetailHandler(settings config.Settings, pool *pgxpool.Pool, registry *payments.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, false); !ok {
			return
		}
		paymentID, err := parsePositiveInt64(chi.URLParam(r, "payment_id"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_payment_id"})
			return
		}
		order, err := registry.Get(r.Context(), paymentID)
		if err != nil {
			status := http.StatusInternalServerError
			code := "payment_load_failed"
			if errors.Is(err, pgx.ErrNoRows) {
				status = http.StatusNotFound
				code = "payment_not_found"
			}
			writeJSON(w, status, map[string]any{"ok": false, "error": code})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "payment": order})
	}
}

func requireAdmin(w http.ResponseWriter, r *http.Request, settings config.Settings, pool *pgxpool.Pool, csrf bool) (sessionContext, bool) {
	session, ok := requireSession(w, r, settings, pool, csrf)
	if !ok {
		return sessionContext{}, false
	}
	if !session.User.IsAdmin {
		writeJSON(w, http.StatusForbidden, map[string]any{"ok": false, "error": "forbidden"})
		return sessionContext{}, false
	}
	return session, true
}

func requireSession(w http.ResponseWriter, r *http.Request, settings config.Settings, pool *pgxpool.Pool, csrf bool) (sessionContext, bool) {
	if pool == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": "database_not_configured"})
		return sessionContext{}, false
	}
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"ok": false, "error": "auth_required"})
		return sessionContext{}, false
	}
	manager := webappSessionManager(settings)
	claims, err := manager.Verify(cookie.Value)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"ok": false, "error": "auth_required"})
		return sessionContext{}, false
	}
	if csrf {
		got := r.Header.Get("X-CSRF-Token")
		csrfCookie, _ := r.Cookie(csrfCookieName)
		if got == "" || got != claims.CSRF || csrfCookie == nil || csrfCookie.Value != claims.CSRF {
			writeJSON(w, http.StatusForbidden, map[string]any{"ok": false, "error": "invalid_csrf"})
			return sessionContext{}, false
		}
	}
	user, err := loadWebappUser(r.Context(), pool, claims.UserID, settings)
	if err != nil {
		status := http.StatusInternalServerError
		code := "user_load_failed"
		if errors.Is(err, pgx.ErrNoRows) {
			status = http.StatusUnauthorized
			code = "auth_required"
		}
		writeJSON(w, status, map[string]any{"ok": false, "error": code})
		return sessionContext{}, false
	}
	return sessionContext{Claims: claims, User: user}, true
}

func webappSessionManager(settings config.Settings) *auth.Manager {
	return auth.NewManager(settings.WebAppSessionSecret, webappSessionSecretFallback(settings))
}

func webappSessionSecretFallback(settings config.Settings) string {
	if strings.TrimSpace(settings.BotToken) != "" {
		return "telegram-bot:" + settings.BotToken
	}
	if strings.TrimSpace(settings.AdminPassword) != "" {
		return "admin-password:" + settings.AdminPassword
	}
	return ""
}

func upsertTelegramUser(ctx context.Context, pool *pgxpool.Pool, user auth.TelegramUser, language string) error {
	// Check whether any user data has changed before writing to avoid redundant WAL.
	var existingUsername, existingFirstName, existingLastName, existingLang, existingPhoto string
	err := pool.QueryRow(ctx, `SELECT COALESCE(username,''), COALESCE(first_name,''), COALESCE(last_name,''),
 COALESCE(language_code,''), COALESCE(telegram_photo_url,'') FROM users WHERE user_id=$1`, user.ID).
		Scan(&existingUsername, &existingFirstName, &existingLastName, &existingLang, &existingPhoto)
	if err == nil {
		if existingUsername == user.Username && existingFirstName == user.FirstName &&
			existingLastName == user.LastName && existingLang == language && existingPhoto == user.PhotoURL {
			return nil
		}
	}

	_, err = pool.Exec(ctx, `
INSERT INTO users (user_id, telegram_id, username, first_name, last_name, language_code, telegram_photo_url, registration_date)
VALUES ($1,$1,$2,$3,$4,$5,$6,NOW())
ON CONFLICT (user_id) DO UPDATE SET
	telegram_id=EXCLUDED.telegram_id,
	username=EXCLUDED.username,
	first_name=EXCLUDED.first_name,
	last_name=EXCLUDED.last_name,
	language_code=EXCLUDED.language_code,
	telegram_photo_url=EXCLUDED.telegram_photo_url`,
		user.ID, emptyStringToNil(user.Username), emptyStringToNil(user.FirstName), emptyStringToNil(user.LastName),
		language, emptyStringToNil(user.PhotoURL))
	return err
}

func loadWebappUser(ctx context.Context, pool *pgxpool.Pool, userID int64, settings config.Settings) (webappUser, error) {
	var user webappUser
	var username, email, firstName, lastName, language, photo, panelUUID string
	err := pool.QueryRow(ctx, `
SELECT user_id, COALESCE(telegram_id,0), COALESCE(username,''), COALESCE(email,''), COALESCE(first_name,''), COALESCE(last_name,''),
	COALESCE(language_code,''), COALESCE(telegram_photo_url,''), COALESCE(panel_user_uuid,'')
FROM users WHERE user_id=$1`, userID).Scan(
		&user.UserID, &user.TelegramID, &username, &email, &firstName, &lastName, &language, &photo, &panelUUID,
	)
	if err != nil {
		return webappUser{}, err
	}
	user.Username = username
	user.Email = email
	user.FirstName = firstName
	user.LastName = lastName
	user.LanguageCode = normalizeWebLanguage(language, effectiveDefaultLanguage(ctx, pool, settings))
	user.PhotoURL = photo
	user.PanelUserUUID = panelUUID
	user.IsAdmin = isAdminID(settings.AdminIDs, user.UserID) || isAdminID(settings.AdminIDs, user.TelegramID)
	return user, nil
}

func setSessionCookies(w http.ResponseWriter, r *http.Request, token string, csrf string) {
	secure := isSecureRequest(r)
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int((30 * 24 * time.Hour).Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     csrfCookieName,
		Value:    csrf,
		Path:     "/",
		MaxAge:   int((30 * 24 * time.Hour).Seconds()),
		HttpOnly: false,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearCookie(w http.ResponseWriter, r *http.Request, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: name == sessionCookieName,
		Secure:   isSecureRequest(r),
		SameSite: http.SameSiteLaxMode,
	})
}

func isSecureRequest(r *http.Request) bool {
	return r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}

func webappFeatureSettings(ctx context.Context, settings config.Settings, pool *pgxpool.Pool, panel *remnawave.Client, user webappUser) map[string]any {
	store := appsettings.NewStore(pool)
	panelConfigured := panel != nil && panel.Configured(ctx)
	trialEnabled := store.Bool(ctx, "TRIAL_ENABLED", false)
	trialDurationDays := store.Int(ctx, "TRIAL_DURATION_DAYS", 0)
	trialTrafficGB := store.Float(ctx, "TRIAL_TRAFFIC_LIMIT_GB", 0)
	referralWelcomeDays := store.Int(ctx, "REFERRAL_WELCOME_BONUS_DAYS", 0)
	myDevicesEnabled := store.Bool(ctx, "MY_DEVICES_ENABLED", panelConfigured)
	supportEnabled := store.Bool(ctx, "SUPPORT_TICKETS_ENABLED", false)
	guidesEnabled := store.Bool(ctx, "SUBSCRIPTION_GUIDES_ENABLED", false)
	autoRenewEnabled := store.Bool(ctx, "SUBSCRIPTION_AUTO_RENEW_ENABLED", false)
	return map[string]any{
		"my_devices_enabled":              myDevicesEnabled,
		"support_tickets_enabled":         supportEnabled,
		"subscription_guides_enabled":     guidesEnabled,
		"trial_enabled":                   trialEnabled,
		"trial_available":                 trialEnabled && trialDurationDays > 0 && panelConfigured && trialAvailableForUser(ctx, pool, user.UserID),
		"trial_duration_days":             trialDurationDays,
		"trial_traffic_limit_gb":          trialTrafficGB,
		"trial_traffic_strategy":          store.String(ctx, "TRIAL_TRAFFIC_STRATEGY", settings.UserTrafficStrategy),
		"trial_requires_telegram":         false,
		"trial_block_reason":              "",
		"referral_welcome_bonus_days":     referralWelcomeDays,
		"tariff_change_enabled":           panelConfigured,
		"subscription_auto_renew_enabled": autoRenewEnabled,
		"support_url":                    store.String(ctx, "SUPPORT_LINK", ""),
		"server_status_url":              store.String(ctx, "SERVER_STATUS_URL", ""),
		"email_auth_enabled":             store.Bool(ctx, "SMTP_ENABLED", false) && mailerConfigFromSettings(ctx, store).IsConfigured(),
	}
}

func paymentStatusPayload(order payments.Order) map[string]any {
	paid := strings.EqualFold(order.Status, "paid") || strings.EqualFold(order.Status, "succeeded")
	return map[string]any{
		"ok":                  true,
		"payment_id":          order.PaymentID,
		"order_id":            order.OrderID,
		"status":              order.Status,
		"paid":                paid,
		"provider":            order.Provider,
		"method":              order.Method,
		"payment_type":        order.PaymentType,
		"payment_url":         order.PaymentURL,
		"qr_content":          order.QRContent,
		"display_amount":      order.DisplayAmount,
		"display_currency":    order.DisplayCurrency,
		"payment_address":     order.PaymentAddress,
		"network":             order.Network,
		"provider_payment_id": order.ProviderPaymentID,
		"provisioned":         order.ProvisionedAt != nil,
		"provisioned_at":      order.ProvisionedAt,
		"provision_error":     order.ProvisionError,
	}
}

func mapSettingFields(ctx context.Context, store appsettings.Store, fields []paymentSettingField) []map[string]any {
	mapped := make([]map[string]any, 0, len(fields))
	for _, field := range fields {
		mapped = append(mapped, field.toMap(ctx, store))
	}
	return mapped
}

func adminSettingsSections(ctx context.Context, settings config.Settings, store appsettings.Store) ([]map[string]any, error) {
	sections := []map[string]any{
		{"id": "general", "fields": mapSettingFields(ctx, store, []paymentSettingField{
			{Key: "DEFAULT_LANGUAGE", Type: "string", Label: "Default language", Description: "Fallback language when a user has no saved preference.", Fallback: settings.DefaultLanguage, Choices: []settingChoice{{Value: "zh", Label: "中文"}, {Value: "en", Label: "English"}}},
			{Key: "TELEGRAM_LOGIN_CLIENT_ID", Type: "string", Label: "Telegram Login Client ID", Description: "Client ID from BotFather Web Login for browser sign-in and account linking.", Fallback: os.Getenv("TELEGRAM_LOGIN_CLIENT_ID")},
			{Key: "SUPPORT_LINK", Type: "string", Label: "Support link", Description: "External support URL shown in the Web App and bot messages.", Fallback: ""},
			{Key: "SERVER_STATUS_URL", Type: "string", Label: "Server status URL", Description: "Optional public service-status page.", Fallback: ""},
			{Key: "PRIVACY_POLICY_URL", Type: "string", Label: "Privacy policy URL", Description: "Optional privacy policy shown in account settings.", Fallback: ""},
			{Key: "USER_AGREEMENT_URL", Type: "string", Label: "User agreement URL", Description: "Optional user agreement shown in account settings.", Fallback: ""},
		})},
		{"id": "remnawave", "fields": remnawaveSettingsFields(ctx, settings, store)},
		{"id": "features", "fields": mapSettingFields(ctx, store, []paymentSettingField{
			{Key: "MY_DEVICES_ENABLED", Type: "bool", Label: "My devices enabled", Description: "Show IP-based device management in the Web App.", Fallback: settings.PanelAPIURL != "" && settings.PanelAPIKey != ""},
			{Key: "SUPPORT_TICKETS_ENABLED", Type: "bool", Label: "Support tickets enabled", Description: "Show the built-in support ticket UI.", Fallback: false},
			{Key: "SUBSCRIPTION_AUTO_RENEW_ENABLED", Type: "bool", Label: "Auto-renew enabled", Description: "Allow users to toggle subscription auto-renewal.", Fallback: false},
		})},
		{"id": "notifications", "fields": mapSettingFields(ctx, store, []paymentSettingField{
			{Key: "SUBSCRIPTION_NOTIFICATIONS_ENABLED", Type: "bool", Label: "Subscription notifications enabled", Description: "Send expiry and traffic notifications through Telegram or email.", Fallback: true},
			{Key: "SUBSCRIPTION_NOTIFY_DAYS_BEFORE", Type: "int", Label: "Notify days before expiry", Description: "Days before subscription expiry to send notifications.", Fallback: settings.SubscriptionNotifyDaysBefore},
			{Key: "SUBSCRIPTION_NOTIFY_HOURS_BEFORE", Type: "int", Label: "Notify hours before expiry", Description: "Additional hours-before notification. 0 disables it.", Fallback: settings.SubscriptionNotifyHoursBefore},
		})},
		{"id": "telemetry", "fields": mapSettingFields(ctx, store, []paymentSettingField{
			{Key: "TELEMETRY_ENABLED", Type: "bool", Label: "Anonymous local telemetry", Description: "Collect local-only installation heartbeats and browser risk signals.", Fallback: envBoolValue("TELEMETRY_ENABLED", true)},
			{Key: "TELEMETRY_RETENTION_HOURS", Type: "int", Label: "Telemetry retention hours", Description: "Delete anonymous records after this many hours without a match (1-720).", Fallback: envIntValue("TELEMETRY_RETENTION_HOURS", 24)},
			{Key: "TELEMETRY_FINGERPRINT_REJECT_SCORE", Type: "int", Label: "Fingerprint rejection score", Description: "Reject welcome bonuses at or above this similarity score (1-100).", Fallback: envIntValue("TELEMETRY_FINGERPRINT_REJECT_SCORE", 70)},
		})},
		{"id": "payments", "fields": paymentSettingsFields(ctx, settings, store)},
		{"id": "mail", "fields": mailSettingsFields(ctx, settings, store)},
		{"id": "subscription_guides", "fields": mapSettingFields(ctx, store, []paymentSettingField{
			{Key: "SUBSCRIPTION_GUIDES_ENABLED", Type: "bool", Label: "Subscription guides enabled", Description: "Show connection guides in the Web App.", Fallback: false},
			{Key: "SUBSCRIPTION_GUIDES_CONFIG", Type: "json", Label: "Subscription guides configuration", Description: "JSON object containing platform and application instructions.", Fallback: map[string]any{}},
		})},
		{"id": "appearance", "fields": mapSettingFields(ctx, store, []paymentSettingField{
			{Key: "WEBAPP_TITLE", Type: "string", Label: "Web App title", Description: "Name displayed in the header and browser tab.", Fallback: "Subscription"},
			{Key: "WEBAPP_LOGO_URL", Type: "string", Label: "Logo URL", Description: "Logo used by the Web App and admin panel.", Fallback: ""},
			{Key: "WEBAPP_FAVICON_USE_CUSTOM", Type: "bool", Label: "Use a separate favicon", Description: "Use the custom favicon URL instead of the logo.", Fallback: false},
			{Key: "WEBAPP_FAVICON_URL", Type: "string", Label: "Favicon URL", Description: "Custom browser icon URL.", Fallback: ""},
		})},
	}
	seen := map[string]bool{}
	for _, section := range sections {
		fields, _ := section["fields"].([]map[string]any)
		for _, field := range fields {
			key, _ := field["key"].(string)
			if key == "" || seen[key] {
				return nil, fmt.Errorf("duplicate setting key %q", key)
			}
			seen[key] = true
		}
	}
	return sections, nil
}

func allowedAdminSettingKeys(sections []map[string]any) map[string]bool {
	result := map[string]bool{}
	for _, section := range sections {
		fields, _ := section["fields"].([]map[string]any)
		for _, field := range fields {
			if key, ok := field["key"].(string); ok && key != "" {
				result[key] = true
			}
		}
	}
	return result
}

func mailSettingsFields(ctx context.Context, settings config.Settings, store appsettings.Store) []map[string]any {
	fields := []paymentSettingField{
		{Key: "SMTP_ENABLED", Type: "bool", Label: "SMTP email enabled", Description: "Enable SMTP email delivery for verification codes, password resets, and notifications.", Subsection: "General", Fallback: false},
		{Key: "SMTP_HOST", Type: "string", Label: "SMTP host", Description: "SMTP server hostname or IP address.", Subsection: "Server", Fallback: ""},
		{Key: "SMTP_PORT", Type: "int", Label: "SMTP port", Description: "SMTP server port (465 for TLS, 587 for STARTTLS, 25 for plain).", Subsection: "Server", Fallback: 587},
		{Key: "SMTP_ENCRYPTION", Type: "string", Label: "SMTP encryption", Description: "Connection encryption method.", Subsection: "Server", Fallback: "starttls", Choices: []settingChoice{{Value: "none", Label: "None"}, {Value: "tls", Label: "TLS (SSL)"}, {Value: "starttls", Label: "STARTTLS"}}},
		{Key: "SMTP_USERNAME", Type: "string", Label: "SMTP username", Description: "SMTP authentication username. Leave empty if no authentication.", Subsection: "Authentication", Fallback: ""},
		{Key: "SMTP_PASSWORD", Type: "string", Label: "SMTP password", Description: "SMTP authentication password.", Subsection: "Authentication", Fallback: "", Secret: true},
		{Key: "SMTP_FROM_EMAIL", Type: "string", Label: "From email", Description: "Sender email address for outgoing messages.", Subsection: "Sender", Fallback: ""},
		{Key: "SMTP_FROM_NAME", Type: "string", Label: "From name", Description: "Sender display name (e.g. your brand name).", Subsection: "Sender", Fallback: ""},
		{Key: "BRAND_NAME", Type: "string", Label: "Brand name", Description: "Brand name used in email subjects and footers.", Subsection: "Sender", Fallback: "Remna"},
		{Key: "EMAIL_TEMPLATE_VERIFY", Type: "text", Label: "Verification email template", Description: "Markdown template for email verification codes. Variables: {{.Brand}}, {{.Code}}, {{.ExpireMinutes}}.", Subsection: "Templates", Fallback: ""},
		{Key: "EMAIL_TEMPLATE_PASSWORD_RESET", Type: "text", Label: "Password reset email template", Description: "Markdown template for password reset codes. Variables: {{.Brand}}, {{.Code}}, {{.ExpireMinutes}}.", Subsection: "Templates", Fallback: ""},
		{Key: "EMAIL_TEMPLATE_LOGIN", Type: "text", Label: "Login code email template", Description: "Markdown template for login verification codes. Variables: {{.Brand}}, {{.Code}}, {{.ExpireMinutes}}.", Subsection: "Templates", Fallback: ""},
	}
	return mapSettingFields(ctx, store, fields)
}

func mailerConfigFromSettings(ctx context.Context, store appsettings.Store) mail.Config {
	return mail.Config{
		Host:       store.String(ctx, "SMTP_HOST", ""),
		Port:       store.Int(ctx, "SMTP_PORT", 587),
		Username:   store.String(ctx, "SMTP_USERNAME", ""),
		Password:   store.String(ctx, "SMTP_PASSWORD", ""),
		FromName:   store.String(ctx, "SMTP_FROM_NAME", ""),
		FromEmail:  store.String(ctx, "SMTP_FROM_EMAIL", ""),
		Encryption: store.String(ctx, "SMTP_ENCRYPTION", "starttls"),
	}
}

func paymentSettingsFields(ctx context.Context, settings config.Settings, store appsettings.Store) []map[string]any {
	effectiveWebhookBaseURL := store.String(ctx, "WEBHOOK_BASE_URL", settings.WebhookBaseURL)
	webhookConfigured := strings.TrimSpace(effectiveWebhookBaseURL) != ""
	fields := []paymentSettingField{
		{Key: "WEBHOOK_BASE_URL", Type: "string", Label: "Webhook base URL", Description: "Public HTTPS backend URL used by payment callbacks.", Subsection: "General", Fallback: settings.WebhookBaseURL},
		{Key: "DEFAULT_CURRENCY_SYMBOL", Type: "string", Label: "Default catalog currency", Description: "Primary catalog currency. New deployments should use USD.", Subsection: "Currency", Fallback: effectiveDefaultCurrency(ctx, settings, nil), Choices: []settingChoice{{Value: "USD", Label: "USD"}, {Value: "CNY", Label: "CNY"}}},
		{Key: "FX_PROVIDER", Type: "string", Label: "USD/CNY rate provider", Description: "Use frankfurter by default; custom uses the fixed rate below.", Subsection: "Currency", Fallback: "frankfurter", Choices: []settingChoice{{Value: "frankfurter", Label: "Frankfurter"}, {Value: "exchange_rate_api", Label: "ExchangeRate-API"}, {Value: "custom", Label: "Custom"}}},
		{Key: "FX_CUSTOM_USD_CNY", Type: "float", Label: "Custom USD/CNY rate", Description: "Used only when the provider is custom.", Subsection: "Currency", Fallback: ""},
		{Key: "FX_CACHE_TTL_SECONDS", Type: "int", Label: "FX cache TTL seconds", Description: "How long a successful rate response is reused.", Subsection: "Currency", Fallback: 3600},
		{Key: "PAYMENT_METHODS_ORDER", Type: "text", Label: "Payment method order", Description: "Comma-separated method ids, e.g. ezpay:alipay,bepusdt:usdt.polygon.", Fallback: strings.Join(settings.PaymentMethodsOrder, ",")},
		{Key: "STARS_ENABLED", Type: "bool", Label: "Telegram Stars enabled", Description: "Accept digital-goods payments through Telegram Stars.", Subsection: "Telegram Stars", Fallback: false},
		{Key: "STARS_USD_RATE", Type: "float", Label: "USD to Stars rate", Description: "Exchange rate: 1 USD = X Stars. Stars price = USD price × this rate.", Subsection: "Telegram Stars", Fallback: settings.StarsUSDRate},
		{Key: "EZPAY_ENABLED", Type: "bool", Label: "EZPay enabled", Description: "Enable EZPay payment methods.", Subsection: "EZPay", Fallback: settings.EZPay.Enabled, WebhookPath: "/webhook/ezpay", ProviderID: "ezpay", WebhookConfigured: webhookConfigured},
		{Key: "EZPAY_BASE_URL", Type: "string", Label: "EZPay base URL", Description: "Merchant API base URL.", Subsection: "EZPay", Fallback: settings.EZPay.BaseURL},
		{Key: "EZPAY_PID", Type: "int", Label: "EZPay PID", Description: "Merchant PID.", Subsection: "EZPay", Fallback: settings.EZPay.PID},
		{Key: "EZPAY_KEY", Type: "string", Label: "EZPay key", Description: "Merchant signing key.", Subsection: "EZPay", Fallback: settings.EZPay.Key, Secret: true},
		{Key: "BEPUSDT_ENABLED", Type: "bool", Label: "BEPUSDT enabled", Description: "Enable BEPUSDT USDT methods.", Subsection: "BEPUSDT", Fallback: settings.BEPUSDT.Enabled, WebhookPath: "/webhook/bepusdt", ProviderID: "bepusdt", WebhookConfigured: webhookConfigured},
		{Key: "BEPUSDT_BASE_URL", Type: "string", Label: "BEPUSDT base URL", Description: "BEPUSDT API base URL.", Subsection: "BEPUSDT", Fallback: settings.BEPUSDT.BaseURL},
		{Key: "BEPUSDT_TOKEN", Type: "string", Label: "BEPUSDT token", Description: "BEPUSDT signing token.", Subsection: "BEPUSDT", Fallback: settings.BEPUSDT.Token, Secret: true},
	}
	result := make([]map[string]any, 0, len(fields))
	for _, field := range fields {
		result = append(result, field.toMap(ctx, store))
	}
	return result
}

func remnawaveSettingsFields(ctx context.Context, settings config.Settings, store appsettings.Store) []map[string]any {
	fields := []paymentSettingField{
		{Key: "PANEL_API_URL", Type: "string", Label: "Remnawave API URL", Description: "Panel base URL. Both https://panel.example and https://panel.example/api are accepted.", Subsection: "Remnawave", Fallback: settings.PanelAPIURL},
		{Key: "PANEL_API_KEY", Type: "string", Label: "Remnawave API key", Description: "Bearer token used to call Remnawave API.", Subsection: "Remnawave", Fallback: settings.PanelAPIKey, Secret: true},
		{Key: "PANEL_API_TOTAL_TIMEOUT_SECONDS", Type: "float", Label: "Remnawave API total timeout", Description: "Maximum total time for one Remnawave API request, in seconds.", Subsection: "Remnawave", Fallback: settings.PanelAPITotalTimeout.Seconds()},
		{Key: "USER_TRAFFIC_LIMIT_GB", Type: "float", Label: "Default traffic limit GB", Description: "Fallback traffic limit when a tariff does not define monthly_gb.", Subsection: "Defaults", Fallback: settings.UserTrafficLimitGB},
		{Key: "USER_TRAFFIC_STRATEGY", Type: "string", Label: "Traffic reset strategy", Description: "Remnawave trafficLimitStrategy for provisioned users.", Subsection: "Defaults", Fallback: settings.UserTrafficStrategy, Choices: []settingChoice{{Value: "NO_RESET", Label: "No reset"}, {Value: "DAY", Label: "Day"}, {Value: "WEEK", Label: "Week"}, {Value: "MONTH", Label: "Month"}, {Value: "MONTH_ROLLING", Label: "Month rolling"}}},
		{Key: "USER_SQUAD_UUIDS", Type: "text", Label: "Default internal squads", Description: "Comma-separated Internal Squad UUIDs used when a tariff has no squad_uuids.", Subsection: "Defaults", Fallback: strings.Join(settings.UserSquadUUIDs, ",")},
		{Key: "USER_EXTERNAL_SQUAD_UUID", Type: "string", Label: "Default external squad", Description: "Optional external squad UUID.", Subsection: "Defaults", Fallback: settings.UserExternalSquadUUID},
		{Key: "TRIAL_ENABLED", Type: "bool", Label: "Trial enabled", Description: "Allow users to activate a one-time trial via Remnawave.", Subsection: "Trial", Fallback: false},
		{Key: "TRIAL_DURATION_DAYS", Type: "int", Label: "Trial duration days", Description: "Number of days granted by trial activation.", Subsection: "Trial", Fallback: 0},
		{Key: "TRIAL_TRAFFIC_LIMIT_GB", Type: "float", Label: "Trial traffic limit GB", Description: "Traffic limit applied to trial users. Empty or 0 falls back to default traffic limit.", Subsection: "Trial", Fallback: 0},
		{Key: "TRIAL_TRAFFIC_STRATEGY", Type: "string", Label: "Trial traffic strategy", Description: "Remnawave trafficLimitStrategy for trial users.", Subsection: "Trial", Fallback: settings.UserTrafficStrategy, Choices: []settingChoice{{Value: "NO_RESET", Label: "No reset"}, {Value: "DAY", Label: "Day"}, {Value: "WEEK", Label: "Week"}, {Value: "MONTH", Label: "Month"}, {Value: "MONTH_ROLLING", Label: "Month rolling"}}},
		{Key: "TRIAL_SQUAD_UUIDS", Type: "text", Label: "Trial internal squads", Description: "Comma-separated Internal Squad UUIDs for trial users. Empty uses default squads.", Subsection: "Trial", Fallback: ""},
		{Key: "REFERRAL_WELCOME_BONUS_DAYS", Type: "int", Label: "Referral welcome bonus days", Description: "Days granted once to invited users. 0 disables welcome bonus claiming.", Subsection: "Referral", Fallback: 0},
	}
	return mapSettingFields(ctx, store, fields)
}

type paymentSettingField struct {
	Key               string
	Type              string
	Label             string
	Description       string
	Subsection        string
	Fallback          any
	Secret            bool
	WebhookPath       string
	ProviderID        string
	WebhookConfigured bool
	Choices           []settingChoice
}

type settingChoice struct {
	Value string
	Label string
}

func (f paymentSettingField) toMap(ctx context.Context, store appsettings.Store) map[string]any {
	raw, overridden, _ := store.Get(ctx, f.Key)
	value := f.Fallback
	envRaw, envLocked := os.LookupEnv(f.Key)
	if envLocked {
		overridden = false
		if parsed, err := normalizeFieldValue(f.Type, envRaw); err == nil {
			value = parsed
		}
	}
	hasValue := false
	if overridden && !envLocked {
		hasValue = true
		switch f.Type {
		case "bool":
			var parsed bool
			if json.Unmarshal(raw, &parsed) == nil {
				value = parsed
			}
		case "int":
			var parsed float64
			if json.Unmarshal(raw, &parsed) == nil {
				value = int(parsed)
			}
		case "float":
			var parsed float64
			if json.Unmarshal(raw, &parsed) == nil {
				value = parsed
			}
		case "json":
			var parsed any
			if json.Unmarshal(raw, &parsed) == nil {
				value = parsed
			}
		default:
			var parsed string
			if json.Unmarshal(raw, &parsed) == nil {
				value = parsed
				hasValue = parsed != ""
			}
		}
	}
	mappedType := f.Type
	if mappedType == "string" {
		mappedType = "input"
	}
	item := map[string]any{
		"key":                  f.Key,
		"type":                 mappedType,
		"label":                f.Label,
		"description":          f.Description,
		"value":                value,
		"overridden":           overridden,
		"env_locked":           envLocked,
		"source":               map[bool]string{true: "env", false: "runtime"}[envLocked],
		"i18n_label_key":       "admin_settings_field_" + strings.ToLower(f.Key) + "_label",
		"i18n_description_key": "admin_settings_field_" + strings.ToLower(f.Key) + "_description",
	}
	if !envLocked && !overridden {
		item["source"] = "default"
	}
	if f.Subsection != "" {
		item["subsection"] = f.Subsection
	}
	if f.Secret {
		item["secret"] = true
		item["has_value"] = hasValue || fmt.Sprint(value) != ""
	}
	if f.WebhookPath != "" {
		item["webhook_path"] = f.WebhookPath
		item["provider_id"] = f.ProviderID
		item["webhook_requires_base_url"] = true
		item["webhook_base_url_configured"] = f.WebhookConfigured
	}
	if len(f.Choices) > 0 {
		choices := make([]map[string]string, 0, len(f.Choices))
		for _, choice := range f.Choices {
			choices = append(choices, map[string]string{"value": choice.Value, "label": choice.Label})
		}
		item["choices"] = choices
	}
	return item
}

func normalizeFieldValue(fieldType, raw string) (any, error) {
	switch fieldType {
	case "bool":
		value, err := strconv.ParseBool(strings.TrimSpace(raw))
		if err != nil {
			return nil, err
		}
		return value, nil
	case "int":
		value, err := strconv.Atoi(strings.TrimSpace(raw))
		if err != nil {
			return nil, err
		}
		return value, nil
	case "float":
		value, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
		if err != nil {
			return nil, err
		}
		return value, nil
	default:
		return strings.TrimSpace(raw), nil
	}
}

func envIntValue(key string, fallback int) int {
	raw, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return fallback
	}
	return value
}
func envBoolValue(key string, fallback bool) bool {
	raw, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	value, err := strconv.ParseBool(strings.TrimSpace(raw))
	if err != nil {
		return fallback
	}
	return value
}

func normalizeSettingValue(key string, value any) (any, error) {
	switch key {
	case "EZPAY_ENABLED", "BEPUSDT_ENABLED", "STARS_ENABLED", "MY_DEVICES_ENABLED", "SUPPORT_TICKETS_ENABLED", "TRIAL_ENABLED", "SUBSCRIPTION_GUIDES_ENABLED", "SUBSCRIPTION_AUTO_RENEW_ENABLED", "SUBSCRIPTION_NOTIFICATIONS_ENABLED", "TELEMETRY_ENABLED", "SMTP_ENABLED", "WEBAPP_FAVICON_USE_CUSTOM":
		if typed, ok := value.(bool); ok {
			return typed, nil
		}
		switch strings.ToLower(strings.TrimSpace(fmt.Sprint(value))) {
		case "true", "1", "yes", "on":
			return true, nil
		case "false", "0", "no", "off":
			return false, nil
		default:
			return nil, fmt.Errorf("invalid_bool")
		}
	case "EZPAY_PID", "FX_CACHE_TTL_SECONDS", "TRIAL_DURATION_DAYS", "REFERRAL_WELCOME_BONUS_DAYS", "TELEMETRY_RETENTION_HOURS", "TELEMETRY_FINGERPRINT_REJECT_SCORE",
		"SUBSCRIPTION_NOTIFY_DAYS_BEFORE", "SUBSCRIPTION_NOTIFY_HOURS_BEFORE",
		"SMTP_PORT":
		var parsed int
		switch typed := value.(type) {
		case float64:
			parsed = int(typed)
		case int:
			parsed = typed
		default:
			value, err := strconv.Atoi(strings.TrimSpace(fmt.Sprint(value)))
			if err != nil {
				return nil, fmt.Errorf("invalid_int")
			}
			parsed = value
		}
		if key == "TELEMETRY_RETENTION_HOURS" && (parsed < 1 || parsed > 720) {
			return nil, fmt.Errorf("must_be_between_1_and_720")
		}
		if key == "TELEMETRY_FINGERPRINT_REJECT_SCORE" && (parsed < 1 || parsed > 100) {
			return nil, fmt.Errorf("must_be_between_1_and_100")
		}
		if key == "SMTP_PORT" && (parsed < 1 || parsed > 65535) {
			return nil, fmt.Errorf("must_be_between_1_and_65535")
		}
		if key == "FX_CACHE_TTL_SECONDS" && parsed < 1 {
			return nil, fmt.Errorf("must_be_positive")
		}
		if parsed < 0 {
			return nil, fmt.Errorf("must_be_non_negative")
		}
		return parsed, nil
	case "DEFAULT_CURRENCY_SYMBOL":
		value := strings.ToUpper(strings.TrimSpace(fmt.Sprint(value)))
		switch value {
		case "USD", "CNY":
			return value, nil
		default:
			return nil, fmt.Errorf("unsupported_currency")
		}
	case "DEFAULT_LANGUAGE":
		value := strings.ToLower(strings.TrimSpace(fmt.Sprint(value)))
		switch value {
		case "zh", "en":
			return value, nil
		default:
			return nil, fmt.Errorf("unsupported_language")
		}
	case "FX_PROVIDER":
		value := strings.ToLower(strings.TrimSpace(fmt.Sprint(value)))
		switch value {
		case "frankfurter", "exchange_rate_api", "custom":
			return value, nil
		default:
			return nil, fmt.Errorf("unsupported_fx_provider")
		}
	case "USER_TRAFFIC_STRATEGY", "TRIAL_TRAFFIC_STRATEGY":
		value := strings.ToUpper(strings.TrimSpace(fmt.Sprint(value)))
		switch value {
		case "NO_RESET", "DAY", "WEEK", "MONTH", "MONTH_ROLLING":
			return value, nil
		default:
			return nil, fmt.Errorf("unsupported_traffic_strategy")
		}
	case "SMTP_ENCRYPTION":
		value := strings.ToLower(strings.TrimSpace(fmt.Sprint(value)))
		switch value {
		case "none", "tls", "starttls":
			return value, nil
		default:
			return nil, fmt.Errorf("unsupported_smtp_encryption")
		}
	case "SUPPORT_LINK", "SERVER_STATUS_URL", "PRIVACY_POLICY_URL", "USER_AGREEMENT_URL", "WEBAPP_LOGO_URL", "WEBAPP_FAVICON_URL", "WEBHOOK_BASE_URL", "PANEL_API_URL", "EZPAY_BASE_URL", "BEPUSDT_BASE_URL":
		text := strings.TrimSpace(fmt.Sprint(value))
		if text == "" {
			return "", nil
		}
		if !safeAppearanceURL(text) {
			return nil, fmt.Errorf("invalid_url")
		}
		return text, nil
	case "FX_CUSTOM_USD_CNY":
		if strings.TrimSpace(fmt.Sprint(value)) == "" {
			return "", nil
		}
		parsed, err := strconv.ParseFloat(strings.TrimSpace(fmt.Sprint(value)), 64)
		if err != nil || parsed <= 0 {
			return nil, fmt.Errorf("invalid_float")
		}
		return strconv.FormatFloat(parsed, 'f', -1, 64), nil
	case "USER_TRAFFIC_LIMIT_GB", "TRIAL_TRAFFIC_LIMIT_GB", "PANEL_API_TOTAL_TIMEOUT_SECONDS", "STARS_USD_RATE":
		if strings.TrimSpace(fmt.Sprint(value)) == "" {
			if strings.HasPrefix(key, "PANEL_API_") {
				return nil, fmt.Errorf("invalid_float")
			}
			return 0, nil
		}
		parsed, err := strconv.ParseFloat(strings.TrimSpace(fmt.Sprint(value)), 64)
		if err != nil || parsed < 0 || (strings.HasPrefix(key, "PANEL_API_") && parsed <= 0) {
			return nil, fmt.Errorf("invalid_float")
		}
		return parsed, nil
	case "SUBSCRIPTION_GUIDES_CONFIG":
		var parsed any
		switch typed := value.(type) {
		case string:
			if len(typed) > 512<<10 || json.Unmarshal([]byte(typed), &parsed) != nil {
				return nil, fmt.Errorf("invalid_json")
			}
		default:
			body, err := json.Marshal(typed)
			if err != nil || len(body) > 512<<10 || json.Unmarshal(body, &parsed) != nil {
				return nil, fmt.Errorf("invalid_json")
			}
		}
		object, ok := parsed.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid_guides_config")
		}
		if !validSubscriptionGuidesConfig(object) {
			return nil, fmt.Errorf("invalid_guides_config")
		}
		return object, nil
	default:
		return strings.TrimSpace(fmt.Sprint(value)), nil
	}
}

func validSubscriptionGuidesConfig(config map[string]any) bool {
	platforms, ok := config["platforms"].(map[string]any)
	if !ok || len(platforms) == 0 {
		return false
	}
	for key, rawPlatform := range platforms {
		if strings.TrimSpace(key) == "" {
			return false
		}
		platform, ok := rawPlatform.(map[string]any)
		if !ok {
			return false
		}
		apps, ok := platform["apps"].([]any)
		if !ok || len(apps) == 0 {
			return false
		}
		for _, rawApp := range apps {
			app, ok := rawApp.(map[string]any)
			if !ok || strings.TrimSpace(fmt.Sprint(app["name"])) == "" {
				return false
			}
			blocks, ok := app["blocks"].([]any)
			if !ok || len(blocks) == 0 {
				return false
			}
			for _, rawBlock := range blocks {
				block, ok := rawBlock.(map[string]any)
				if !ok {
					return false
				}
				buttons, _ := block["buttons"].([]any)
				for _, rawButton := range buttons {
					button, ok := rawButton.(map[string]any)
					if !ok {
						return false
					}
					link := strings.TrimSpace(fmt.Sprint(button["link"]))
					lower := strings.ToLower(link)
					if link == "" || strings.ContainsAny(link, "\r\n\x00") || strings.HasPrefix(lower, "javascript:") || strings.HasPrefix(lower, "data:") || strings.HasPrefix(lower, "vbscript:") {
						return false
					}
				}
			}
		}
	}
	return true
}

func paymentErrorCode(err error) string {
	text := strings.ToLower(err.Error())
	switch {
	case strings.Contains(text, "disabled"):
		return "provider_disabled"
	case strings.Contains(text, "not configured"):
		return "provider_not_configured"
	case strings.Contains(text, "webhook"):
		return "webhook_base_url_required"
	case strings.Contains(text, "unsupported"):
		return "payment_method_not_supported"
	default:
		return "payment_create_failed"
	}
}

func normalizeWebLanguage(raw string, fallback string) string {
	value := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(raw), "_", "-"))
	if strings.HasPrefix(value, "zh") {
		return "zh"
	}
	if strings.HasPrefix(value, "en") {
		return "en"
	}
	fallback = strings.ToLower(strings.TrimSpace(fallback))
	if fallback == "en" {
		return "en"
	}
	return "zh"
}

func isAdminID(adminIDs []int64, id int64) bool {
	if id == 0 {
		return false
	}
	for _, adminID := range adminIDs {
		if adminID == id {
			return true
		}
	}
	return false
}

func clientIP(r *http.Request) string {
	for _, header := range []string{"X-Forwarded-For", "X-Real-IP"} {
		value := strings.TrimSpace(r.Header.Get(header))
		if value == "" {
			continue
		}
		if first := strings.TrimSpace(strings.Split(value, ",")[0]); first != "" {
			return first
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

func emptyStringToNil(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func parsePositiveInt64(raw string) (int64, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid_id")
	}
	return id, nil
}

func compactFloat(value float64) string {
	if value == float64(int64(value)) {
		return strconv.FormatInt(int64(value), 10)
	}
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func effectiveDefaultCurrency(ctx context.Context, settings config.Settings, pool *pgxpool.Pool) string {
	value := appsettings.NewStore(pool).String(ctx, "DEFAULT_CURRENCY_SYMBOL", settings.DefaultCurrency)
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "" || value == "RUB" {
		return "USD"
	}
	return value
}

// effectiveDefaultLanguage 从 app_settings 读取默认语言，回退到 env。
func effectiveDefaultLanguage(ctx context.Context, pool *pgxpool.Pool, settings config.Settings) string {
	value := appsettings.NewStore(pool).String(ctx, "DEFAULT_LANGUAGE", settings.DefaultLanguage)
	return normalizeWebLanguage(value, settings.DefaultLanguage)
}

func providerCheckoutAmount(methodID string, plan tariffs.Plan, rate fx.Rate) (float64, string, error) {
	provider, _, ok := strings.Cut(strings.TrimSpace(methodID), ":")
	if !ok || provider == "" {
		return 0, "", fmt.Errorf("invalid payment method")
	}
	baseAmount := plan.BaseAmount
	if baseAmount <= 0 {
		baseAmount = plan.Price
	}
	baseCurrency := strings.ToUpper(strings.TrimSpace(plan.BaseCurrency))
	if baseCurrency == "" {
		baseCurrency = strings.ToUpper(plan.Currency)
	}
	switch provider {
	case payments.ProviderEZPay:
		switch baseCurrency {
		case "CNY", "RMB":
			return roundMoney(baseAmount), "CNY", nil
		case "USD":
			return roundMoney(baseAmount * rate.Rate), "CNY", nil
		default:
			return roundMoney(plan.Price), strings.ToUpper(plan.Currency), nil
		}
	case payments.ProviderBEPUSDT:
		switch baseCurrency {
		case "USD":
			return roundMoney(baseAmount), "USD", nil
		case "CNY", "RMB":
			if rate.Rate <= 0 {
				return 0, "", fmt.Errorf("missing fx rate")
			}
			return roundMoney(baseAmount / rate.Rate), "USD", nil
		default:
			if strings.ToUpper(plan.Currency) == "USD" {
				return roundMoney(plan.Price), "USD", nil
			}
			return 0, "", fmt.Errorf("unsupported currency")
		}
	case payments.ProviderTelegramStars:
		if plan.StarsPrice <= 0 {
			return 0, "", fmt.Errorf("stars price not configured")
		}
		return float64(plan.StarsPrice), "XTR", nil
	default:
		return 0, "", fmt.Errorf("unsupported provider")
	}
}

func roundMoney(value float64) float64 {
	return math.Round(value*100) / 100
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
