package telegram

import (
	"crypto/hmac"
	"crypto/sha256"
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

var (
	// ErrInitDataMalformed means the Mini App data could not be parsed safely.
	ErrInitDataMalformed = errors.New("telegram init data is malformed")
	// ErrInitDataSignature means the Mini App data signature did not verify.
	ErrInitDataSignature = errors.New("telegram init data signature is invalid")
	// ErrInitDataExpired means auth_date is outside the accepted freshness window.
	ErrInitDataExpired = errors.New("telegram init data is expired")
)

// InitData is authenticated Telegram Mini App launch data.
type InitData struct {
	QueryID      string
	User         User
	AuthDate     time.Time
	StartParam   string
	ChatType     string
	ChatInstance string
	RawFields    map[string]string
}

// InitDataVerifier validates Telegram Mini App initData with a bot token.
type InitDataVerifier struct {
	botToken string
	maxAge   time.Duration
	now      func() time.Time
}

// NewInitDataVerifier constructs a verifier. maxAge must be positive.
func NewInitDataVerifier(botToken string, maxAge time.Duration) (*InitDataVerifier, error) {
	if strings.TrimSpace(botToken) == "" {
		return nil, fmt.Errorf("%w: bot token is empty", ErrInitDataMalformed)
	}
	if maxAge <= 0 {
		return nil, fmt.Errorf("%w: max age must be positive", ErrInitDataMalformed)
	}
	return &InitDataVerifier{botToken: botToken, maxAge: maxAge, now: time.Now}, nil
}

// Verify authenticates and parses raw Telegram.WebApp.initData.
func (v *InitDataVerifier) Verify(raw string) (*InitData, error) {
	if v == nil || v.now == nil || v.botToken == "" {
		return nil, fmt.Errorf("%w: verifier is not configured", ErrInitDataMalformed)
	}
	if raw == "" {
		return nil, fmt.Errorf("%w: value is empty", ErrInitDataMalformed)
	}

	values, err := url.ParseQuery(raw)
	if err != nil {
		return nil, fmt.Errorf("%w: parse query: %v", ErrInitDataMalformed, err)
	}
	fields := make(map[string]string, len(values))
	for key, entries := range values {
		if len(entries) != 1 {
			return nil, fmt.Errorf("%w: duplicate field %q", ErrInitDataMalformed, key)
		}
		fields[key] = entries[0]
	}

	providedHash, ok := fields["hash"]
	if !ok || providedHash == "" {
		return nil, fmt.Errorf("%w: hash is missing", ErrInitDataMalformed)
	}
	providedMAC, err := hex.DecodeString(providedHash)
	if err != nil || len(providedMAC) != sha256.Size {
		return nil, fmt.Errorf("%w: hash is not a SHA-256 digest", ErrInitDataMalformed)
	}

	keys := make([]string, 0, len(fields)-1)
	for key := range fields {
		// Bot-token HMAC validation excludes only hash. Telegram's newer
		// signature field is itself part of this data-check-string; only the
		// separate third-party Ed25519 algorithm excludes both fields.
		if key != "hash" {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+fields[key])
	}
	dataCheckString := strings.Join(parts, "\n")

	secretMAC := hmac.New(sha256.New, []byte("WebAppData"))
	secretMAC.Write([]byte(v.botToken))
	checkMAC := hmac.New(sha256.New, secretMAC.Sum(nil))
	checkMAC.Write([]byte(dataCheckString))
	if !hmac.Equal(providedMAC, checkMAC.Sum(nil)) {
		return nil, ErrInitDataSignature
	}

	authUnix, err := strconv.ParseInt(fields["auth_date"], 10, 64)
	if err != nil || authUnix <= 0 {
		return nil, fmt.Errorf("%w: auth_date is invalid", ErrInitDataMalformed)
	}
	authDate := time.Unix(authUnix, 0).UTC()
	age := v.now().UTC().Sub(authDate)
	if age > v.maxAge {
		return nil, ErrInitDataExpired
	}
	if age < -30*time.Second {
		return nil, fmt.Errorf("%w: auth_date is in the future", ErrInitDataMalformed)
	}

	var user User
	if fields["user"] == "" {
		return nil, fmt.Errorf("%w: user is missing", ErrInitDataMalformed)
	}
	if err := json.Unmarshal([]byte(fields["user"]), &user); err != nil {
		return nil, fmt.Errorf("%w: decode user: %v", ErrInitDataMalformed, err)
	}
	if user.ID <= 0 {
		return nil, fmt.Errorf("%w: user id is invalid", ErrInitDataMalformed)
	}

	cleanFields := make(map[string]string, len(fields)-1)
	for key, value := range fields {
		if key != "hash" && key != "signature" {
			cleanFields[key] = value
		}
	}
	return &InitData{
		QueryID:      fields["query_id"],
		User:         user,
		AuthDate:     authDate,
		StartParam:   fields["start_param"],
		ChatType:     fields["chat_type"],
		ChatInstance: fields["chat_instance"],
		RawFields:    cleanFields,
	}, nil
}
