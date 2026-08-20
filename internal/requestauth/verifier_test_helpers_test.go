package requestauth

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

func newTestVerifier(t *testing.T) *Verifier {
	t.Helper()
	verifier, err := New([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	verifier.now = func() time.Time { return time.Unix(1_800_000_000, 0) }
	return verifier
}

func signedRequest(t *testing.T, verifier *Verifier, method, target, body, nonce string, at time.Time) *http.Request {
	return signedRequestForSession(t, verifier, testSessionToken, method, target, body, nonce, at)
}

func signedRequestForSession(t *testing.T, verifier *Verifier, sessionToken, method, target, body, nonce string, at time.Time) *http.Request {
	t.Helper()
	request := httptest.NewRequest(method, target, strings.NewReader(body))
	key, err := verifier.ClientKey(sessionToken)
	if err != nil {
		t.Fatalf("ClientKey(): %v", err)
	}
	timestamp := strconv.FormatInt(at.Unix(), 10)
	signature, err := Sign(key, method, CanonicalTarget(request.URL), timestamp, nonce, []byte(body))
	if err != nil {
		t.Fatalf("Sign(): %v", err)
	}
	request.Header.Set(TimestampHeader, timestamp)
	request.Header.Set(NonceHeader, nonce)
	request.Header.Set(SignatureHeader, signature)
	return request
}
