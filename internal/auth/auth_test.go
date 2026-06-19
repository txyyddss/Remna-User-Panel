package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestValidateTelegramInitData(t *testing.T) {
	botToken := "123456:ABCDEF"
	values := url.Values{}
	values.Set("auth_date", "1893456000")
	values.Set("query_id", "test-query")
	values.Set("user", `{"id":42,"first_name":"Ada","username":"ada","language_code":"zh"}`)
	values.Set("hash", telegramInitDataHash(values, botToken))

	user, err := ValidateTelegramInitData(values.Encode(), botToken, 0)
	if err != nil {
		t.Fatalf("ValidateTelegramInitData returned error: %v", err)
	}
	if user.ID != 42 || user.Username != "ada" {
		t.Fatalf("unexpected user: %#v", user)
	}
}

func TestValidateTelegramInitDataRejectsInvalidHash(t *testing.T) {
	values := url.Values{}
	values.Set("auth_date", "1893456000")
	values.Set("user", `{"id":42}`)
	values.Set("hash", "bad")

	if _, err := ValidateTelegramInitData(values.Encode(), "123456:ABCDEF", time.Hour); err == nil {
		t.Fatal("expected invalid hash error")
	}
}

func TestValidateTelegramAuthData(t *testing.T) {
	botToken := "123456:ABCDEF"
	values := map[string]any{
		"id":         "42",
		"first_name": "Ada",
		"username":   "ada",
		"auth_date":  "1893456000",
	}
	values["hash"] = telegramAuthDataHash(values, botToken)

	user, err := ValidateTelegramAuthData(values, botToken, 0)
	if err != nil {
		t.Fatalf("ValidateTelegramAuthData returned error: %v", err)
	}
	if user.ID != 42 || user.Username != "ada" {
		t.Fatalf("unexpected user: %#v", user)
	}
}

func TestValidateTelegramAuthDataRejectsInvalidHash(t *testing.T) {
	values := map[string]any{
		"id":        "42",
		"auth_date": "1893456000",
		"hash":      "bad",
	}
	if _, err := ValidateTelegramAuthData(values, "123456:ABCDEF", time.Hour); err == nil {
		t.Fatal("expected invalid hash error")
	}
}

func telegramInitDataHash(values url.Values, botToken string) string {
	pairs := make([]string, 0, len(values))
	for key, bucket := range values {
		if key == "hash" || key == "signature" || len(bucket) == 0 {
			continue
		}
		pairs = append(pairs, key+"="+bucket[0])
	}
	sort.Strings(pairs)
	secretMAC := hmac.New(sha256.New, []byte("WebAppData"))
	secretMAC.Write([]byte(botToken))
	dataMAC := hmac.New(sha256.New, secretMAC.Sum(nil))
	dataMAC.Write([]byte(strings.Join(pairs, "\n")))
	return hex.EncodeToString(dataMAC.Sum(nil))
}

func telegramAuthDataHash(values map[string]any, botToken string) string {
	pairs := make([]string, 0, len(values))
	for key, value := range values {
		if key == "hash" {
			continue
		}
		pairs = append(pairs, key+"="+strings.TrimSpace(value.(string)))
	}
	sort.Strings(pairs)
	secret := sha256.Sum256([]byte(botToken))
	dataMAC := hmac.New(sha256.New, secret[:])
	dataMAC.Write([]byte(strings.Join(pairs, "\n")))
	return hex.EncodeToString(dataMAC.Sum(nil))
}
