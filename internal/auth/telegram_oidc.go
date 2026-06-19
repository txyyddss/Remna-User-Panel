package auth

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const telegramJWKSURL = "https://oauth.telegram.org/.well-known/jwks.json"

type telegramOIDCClaims struct {
	Issuer    string `json:"iss"`
	Audience  any    `json:"aud"`
	Subject   string `json:"sub"`
	ExpiresAt int64  `json:"exp"`
	IssuedAt  int64  `json:"iat"`
	Nonce     string `json:"nonce"`
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Username  string `json:"preferred_username"`
	Picture   string `json:"picture"`
}

var telegramKeys = struct {
	sync.Mutex
	keys    map[string]*rsa.PublicKey
	expires time.Time
}{}

// ValidateTelegramIDToken verifies a Telegram Login Library OIDC ID token.
func ValidateTelegramIDToken(ctx context.Context, token, clientID, nonce string) (TelegramUser, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return TelegramUser{}, errors.New("invalid_id_token")
	}
	decode := func(value string) ([]byte, error) { return base64.RawURLEncoding.DecodeString(value) }
	headerBody, err := decode(parts[0])
	if err != nil {
		return TelegramUser{}, err
	}
	var header struct {
		Algorithm string `json:"alg"`
		KeyID     string `json:"kid"`
	}
	if json.Unmarshal(headerBody, &header) != nil || header.Algorithm != "RS256" || header.KeyID == "" {
		return TelegramUser{}, errors.New("invalid_id_token_header")
	}
	claimsBody, err := decode(parts[1])
	if err != nil {
		return TelegramUser{}, err
	}
	var claims telegramOIDCClaims
	if json.Unmarshal(claimsBody, &claims) != nil {
		return TelegramUser{}, errors.New("invalid_id_token_claims")
	}
	key, err := telegramPublicKey(ctx, header.KeyID)
	if err != nil {
		return TelegramUser{}, err
	}
	signature, err := decode(parts[2])
	if err != nil {
		return TelegramUser{}, err
	}
	digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if rsa.VerifyPKCS1v15(key, crypto.SHA256, digest[:], signature) != nil {
		return TelegramUser{}, errors.New("invalid_id_token_signature")
	}
	now := time.Now().Unix()
	if claims.Issuer != "https://oauth.telegram.org" || claims.ExpiresAt <= now || claims.IssuedAt > now+60 || claims.Nonce != nonce || !audienceMatches(claims.Audience, clientID) {
		return TelegramUser{}, errors.New("invalid_id_token_claims")
	}
	id := claims.ID
	if id == 0 {
		id, _ = strconv.ParseInt(claims.Subject, 10, 64)
	}
	if id <= 0 {
		return TelegramUser{}, errors.New("invalid_id_token_user")
	}
	nameParts := strings.Fields(claims.Name)
	first, last := "", ""
	if len(nameParts) > 0 {
		first = nameParts[0]
	}
	if len(nameParts) > 1 {
		last = strings.Join(nameParts[1:], " ")
	}
	return TelegramUser{ID: id, FirstName: first, LastName: last, Username: claims.Username, PhotoURL: claims.Picture}, nil
}

func audienceMatches(raw any, clientID string) bool {
	want := strings.TrimSpace(clientID)
	switch value := raw.(type) {
	case string:
		return value == want
	case []any:
		for _, item := range value {
			if fmt.Sprint(item) == want {
				return true
			}
		}
	}
	return false
}

func telegramPublicKey(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	telegramKeys.Lock()
	defer telegramKeys.Unlock()
	if time.Now().Before(telegramKeys.expires) && telegramKeys.keys[kid] != nil {
		return telegramKeys.keys[kid], nil
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, telegramJWKSURL, nil)
	client := &http.Client{Timeout: 8 * time.Second}
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode/100 != 2 {
		return nil, fmt.Errorf("telegram_jwks_status_%d", response.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	var set struct {
		Keys []struct {
			KeyID string `json:"kid"`
			Type  string `json:"kty"`
			N     string `json:"n"`
			E     string `json:"e"`
		} `json:"keys"`
	}
	if json.Unmarshal(body, &set) != nil {
		return nil, errors.New("invalid_telegram_jwks")
	}
	keys := map[string]*rsa.PublicKey{}
	for _, item := range set.Keys {
		if item.Type != "RSA" {
			continue
		}
		nBytes, nErr := base64.RawURLEncoding.DecodeString(item.N)
		eBytes, eErr := base64.RawURLEncoding.DecodeString(item.E)
		if nErr != nil || eErr != nil || len(eBytes) == 0 || len(eBytes) > 4 {
			continue
		}
		padded := make([]byte, 4)
		copy(padded[4-len(eBytes):], eBytes)
		exponent := int(binary.BigEndian.Uint32(padded))
		keys[item.KeyID] = &rsa.PublicKey{N: new(big.Int).SetBytes(nBytes), E: exponent}
	}
	telegramKeys.keys = keys
	telegramKeys.expires = time.Now().Add(time.Hour)
	key := keys[kid]
	if key == nil {
		return nil, errors.New("telegram_jwks_key_not_found")
	}
	return key, nil
}
