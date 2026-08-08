package telegram

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestInitDataVerifier(t *testing.T) {
	t.Parallel()
	const token = "123456:TEST_TOKEN"
	now := time.Date(2026, time.August, 7, 12, 0, 0, 0, time.UTC)
	base := url.Values{
		"auth_date": {"1786103880"},
		"query_id":  {"AAHdF6IQAAAAAN0XohDhrOrc"},
		"user":      {`{"id":279058397,"first_name":"Ada","username":"ada"}`},
	}
	base.Set("auth_date", strconv.FormatInt(now.Add(-time.Minute).Unix(), 10))
	valid := signedInitData(base, token)

	tests := []struct {
		name    string
		raw     string
		wantErr error
	}{
		{name: "valid", raw: valid},
		{name: "tampered", raw: strings.Replace(valid, "%22Ada%22", "%22Eve%22", 1), wantErr: ErrInitDataSignature},
		{name: "expired", raw: signedInitData(withValue(base, "auth_date", strconv.FormatInt(now.Add(-6*time.Minute).Unix(), 10)), token), wantErr: ErrInitDataExpired},
		{name: "far future", raw: signedInitData(withValue(base, "auth_date", strconv.FormatInt(now.Add(6*time.Minute).Unix(), 10)), token), wantErr: ErrInitDataMalformed},
		{name: "duplicate field", raw: valid + "&auth_date=1", wantErr: ErrInitDataMalformed},
		{name: "bad hash encoding", raw: strings.Replace(valid, "hash=", "hash=xyz", 1), wantErr: ErrInitDataMalformed},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			verifier, err := NewInitDataVerifier(token, 5*time.Minute)
			if err != nil {
				t.Fatalf("NewInitDataVerifier() error = %v", err)
			}
			verifier.now = func() time.Time { return now }
			got, err := verifier.Verify(test.raw)
			if test.wantErr != nil {
				if !errors.Is(err, test.wantErr) {
					t.Fatalf("Verify() error = %v, want errors.Is(%v)", err, test.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Verify() error = %v", err)
			}
			if got.User.ID != 279058397 || got.User.Username != "ada" || got.QueryID == "" {
				t.Fatalf("Verify() = %#v", got)
			}
			if _, exposed := got.RawFields["hash"]; exposed {
				t.Fatal("Verify() retained the authentication hash")
			}
		})
	}
}

func TestNewInitDataVerifierValidation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		token  string
		maxAge time.Duration
	}{
		{name: "empty token", maxAge: time.Minute},
		{name: "zero max age", token: "token"},
		{name: "negative max age", token: "token", maxAge: -time.Second},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if _, err := NewInitDataVerifier(test.token, test.maxAge); err == nil {
				t.Fatal("NewInitDataVerifier() error = nil")
			}
		})
	}
}

func TestInitDataVerifierAcceptsTelegramSignatureField(t *testing.T) {
	t.Parallel()

	const token = "7342037359:AAHI25ES9xCOMPokpYoz-p8XVrZUdygo2J4"
	const raw = "user=%7B%22id%22%3A279058397%2C%22first_name%22%3A%22Vladislav%20%2B%20-%20%3F%20%5C%2F%22%2C%22last_name%22%3A%22Kibenko%22%2C%22username%22%3A%22vdkfrost%22%2C%22language_code%22%3A%22ru%22%2C%22is_premium%22%3Atrue%2C%22allows_write_to_pm%22%3Atrue%2C%22photo_url%22%3A%22https%3A%5C%2F%5C%2Ft.me%5C%2Fi%5C%2Fuserpic%5C%2F320%5C%2F4FPEE4tmP3ATHa57u6MqTDih13LTOiMoKoLDRG4PnSA.svg%22%7D&chat_instance=8134722200314281151&chat_type=private&auth_date=1733509682&signature=TYJxVcisqbWjtodPepiJ6ghziUL94-KNpG8Pau-X7oNNLNBM72APCpi_RKiUlBvcqo5L-LAxIc3dnTzcZX_PDg&hash=a433d8f9847bd6addcc563bff7cc82c89e97ea0d90c11fe5729cae6796a36d73"
	verifier, err := NewInitDataVerifier(token, 5*time.Minute)
	if err != nil {
		t.Fatalf("NewInitDataVerifier(): %v", err)
	}
	verifier.now = func() time.Time { return time.Unix(1733509682, 0).Add(time.Minute) }
	data, err := verifier.Verify(raw)
	if err != nil {
		t.Fatalf("Verify(): %v", err)
	}
	if data.User.ID != 279058397 {
		t.Fatalf("user ID = %d", data.User.ID)
	}
	if _, exposed := data.RawFields["signature"]; exposed {
		t.Fatal("Verify() retained the third-party signature")
	}
}

func signedInitData(values url.Values, token string) string {
	copyValues := make(url.Values, len(values)+1)
	keys := make([]string, 0, len(values))
	for key, entries := range values {
		copyValues[key] = append([]string(nil), entries...)
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+copyValues.Get(key))
	}
	secret := hmac.New(sha256.New, []byte("WebAppData"))
	secret.Write([]byte(token))
	mac := hmac.New(sha256.New, secret.Sum(nil))
	mac.Write([]byte(strings.Join(parts, "\n")))
	copyValues.Set("hash", hex.EncodeToString(mac.Sum(nil)))
	return copyValues.Encode()
}

func withValue(values url.Values, key, value string) url.Values {
	copyValues := make(url.Values, len(values))
	for name, entries := range values {
		copyValues[name] = append([]string(nil), entries...)
	}
	copyValues.Set(key, value)
	return copyValues
}
