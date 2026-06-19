// Package auth validates Telegram Mini App launch data and signs web sessions.
package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	sessionTTL = 30 * 24 * time.Hour
	csrfBytes  = 32
)

// TelegramUser is the signed user object embedded in Telegram initData.
type TelegramUser struct {
	ID           int64  `json:"id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Username     string `json:"username"`
	LanguageCode string `json:"language_code"`
	PhotoURL     string `json:"photo_url"`
}

// SessionClaims are the trusted values stored in a signed session cookie.
type SessionClaims struct {
	UserID int64  `json:"uid"`
	CSRF   string `json:"csrf"`
	Exp    int64  `json:"exp"`
}

// Manager signs and verifies stateless web sessions.
type Manager struct {
	secret     []byte
	configured bool
	now        func() time.Time
}

// NewManager creates a session manager. The fallback is only used when secret is empty.
func NewManager(secret string, fallback string) *Manager {
	value := strings.TrimSpace(secret)
	if value == "" {
		value = strings.TrimSpace(fallback)
	}
	return &Manager{secret: []byte(value), configured: value != "", now: time.Now}
}

// Sign creates a signed session token and returns it with the CSRF value.
func (m *Manager) Sign(userID int64) (string, string, error) {
	if !m.configured {
		return "", "", errors.New("session_secret_not_configured")
	}
	csrf, err := randomToken(csrfBytes)
	if err != nil {
		return "", "", err
	}
	claims := SessionClaims{
		UserID: userID,
		CSRF:   csrf,
		Exp:    m.now().Add(sessionTTL).Unix(),
	}
	body, err := json.Marshal(claims)
	if err != nil {
		return "", "", err
	}
	payload := base64.RawURLEncoding.EncodeToString(body)
	signature := m.sign(payload)
	return payload + "." + signature, csrf, nil
}

// Verify validates a session token and returns its claims.
func (m *Manager) Verify(token string) (SessionClaims, error) {
	if !m.configured {
		return SessionClaims{}, errors.New("session_secret_not_configured")
	}
	parts := strings.Split(token, ".")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return SessionClaims{}, errors.New("invalid_session")
	}
	expected := m.sign(parts[0])
	if subtle.ConstantTimeCompare([]byte(expected), []byte(parts[1])) != 1 {
		return SessionClaims{}, errors.New("invalid_session_signature")
	}
	body, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return SessionClaims{}, fmt.Errorf("decode session: %w", err)
	}
	var claims SessionClaims
	if err := json.Unmarshal(body, &claims); err != nil {
		return SessionClaims{}, fmt.Errorf("parse session: %w", err)
	}
	if claims.UserID == 0 || claims.CSRF == "" || claims.Exp <= m.now().Unix() {
		return SessionClaims{}, errors.New("expired_session")
	}
	return claims, nil
}

func (m *Manager) sign(payload string) string {
	mac := hmac.New(sha256.New, m.secret)
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// ValidateTelegramInitData verifies Telegram Mini App initData and returns the signed user.
func ValidateTelegramInitData(initData string, botToken string, maxAge time.Duration) (TelegramUser, error) {
	if strings.TrimSpace(botToken) == "" {
		return TelegramUser{}, errors.New("telegram_bot_token_not_configured")
	}
	values, err := url.ParseQuery(initData)
	if err != nil {
		return TelegramUser{}, fmt.Errorf("parse init data: %w", err)
	}
	gotHash := values.Get("hash")
	if gotHash == "" {
		return TelegramUser{}, errors.New("missing_hash")
	}
	pairs := make([]string, 0, len(values))
	for key, bucket := range values {
		if key == "hash" || key == "signature" || len(bucket) == 0 {
			continue
		}
		pairs = append(pairs, key+"="+bucket[0])
	}
	sort.Strings(pairs)
	checkString := strings.Join(pairs, "\n")

	secretMAC := hmac.New(sha256.New, []byte("WebAppData"))
	secretMAC.Write([]byte(botToken))
	secret := secretMAC.Sum(nil)
	dataMAC := hmac.New(sha256.New, secret)
	dataMAC.Write([]byte(checkString))
	expectedHash := hex.EncodeToString(dataMAC.Sum(nil))
	if subtle.ConstantTimeCompare([]byte(expectedHash), []byte(gotHash)) != 1 {
		return TelegramUser{}, errors.New("invalid_hash")
	}
	authDate, err := strconv.ParseInt(values.Get("auth_date"), 10, 64)
	if err != nil || authDate <= 0 {
		return TelegramUser{}, errors.New("invalid_auth_date")
	}
	if maxAge > 0 && time.Since(time.Unix(authDate, 0)) > maxAge {
		return TelegramUser{}, errors.New("expired_init_data")
	}
	var user TelegramUser
	if err := json.Unmarshal([]byte(values.Get("user")), &user); err != nil {
		return TelegramUser{}, fmt.Errorf("parse telegram user: %w", err)
	}
	if user.ID == 0 {
		return TelegramUser{}, errors.New("missing_user")
	}
	return user, nil
}

// ValidateTelegramAuthData verifies Telegram Login Widget auth data.
func ValidateTelegramAuthData(authData map[string]any, botToken string, maxAge time.Duration) (TelegramUser, error) {
	if strings.TrimSpace(botToken) == "" {
		return TelegramUser{}, errors.New("telegram_bot_token_not_configured")
	}
	values := map[string]string{}
	for key, value := range authData {
		if key == "" || value == nil {
			continue
		}
		values[key] = strings.TrimSpace(fmt.Sprint(value))
	}
	gotHash := values["hash"]
	if gotHash == "" {
		return TelegramUser{}, errors.New("missing_hash")
	}
	pairs := make([]string, 0, len(values))
	for key, value := range values {
		if key == "hash" || value == "" {
			continue
		}
		pairs = append(pairs, key+"="+value)
	}
	sort.Strings(pairs)
	secret := sha256.Sum256([]byte(botToken))
	dataMAC := hmac.New(sha256.New, secret[:])
	dataMAC.Write([]byte(strings.Join(pairs, "\n")))
	expectedHash := hex.EncodeToString(dataMAC.Sum(nil))
	if subtle.ConstantTimeCompare([]byte(expectedHash), []byte(gotHash)) != 1 {
		return TelegramUser{}, errors.New("invalid_hash")
	}
	authDate, err := strconv.ParseInt(values["auth_date"], 10, 64)
	if err != nil || authDate <= 0 {
		return TelegramUser{}, errors.New("invalid_auth_date")
	}
	if maxAge > 0 && time.Since(time.Unix(authDate, 0)) > maxAge {
		return TelegramUser{}, errors.New("expired_auth_data")
	}
	id, err := strconv.ParseInt(values["id"], 10, 64)
	if err != nil || id == 0 {
		return TelegramUser{}, errors.New("missing_user")
	}
	return TelegramUser{
		ID:        id,
		FirstName: values["first_name"],
		LastName:  values["last_name"],
		Username:  values["username"],
		PhotoURL:  values["photo_url"],
	}, nil
}

func randomToken(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
