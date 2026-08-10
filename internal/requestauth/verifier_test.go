package requestauth

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

const testSessionToken = "c2Vzc2lvbi10b2tlbi0zMi1ieXRlcy0wMDAwMDAwMDA"

func TestVerifyAcceptsCanonicalRequestAndRestoresBody(t *testing.T) {
	t.Parallel()
	verifier := newTestVerifier(t)
	body := `{"name":"Ada","password":" 密碼 päss 🧩 "}`
	request := signedRequest(t, verifier, "POST", "/api/v1/profile?view=full%20name", body, "AAAAAAAAAAAAAAAAAAAAAA", time.Unix(1_800_000_000, 0))

	key, err := verifier.ClientKey(testSessionToken)
	if err != nil {
		t.Fatalf("ClientKey(): %v", err)
	}
	if err := verifier.Verify(request, testSessionToken, key); err != nil {
		t.Fatalf("Verify(): %v", err)
	}
	restored, err := io.ReadAll(request.Body)
	if err != nil {
		t.Fatalf("read restored body: %v", err)
	}
	if string(restored) != body {
		t.Fatalf("restored body = %q, want %q", restored, body)
	}
}

func TestVerifyRejectsInvalidSigningState(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		mutate  func(*httptest.ResponseRecorder, *http.Request, *Verifier, string)
		wantErr error
	}{
		{name: "missing signature", mutate: func(_ *httptest.ResponseRecorder, request *http.Request, _ *Verifier, _ string) {
			request.Header.Del(SignatureHeader)
		}, wantErr: ErrRequired},
		{name: "malformed nonce", mutate: func(_ *httptest.ResponseRecorder, request *http.Request, _ *Verifier, _ string) {
			request.Header.Set(NonceHeader, "too-short")
		}, wantErr: ErrMalformed},
		{name: "stale timestamp", mutate: func(_ *httptest.ResponseRecorder, request *http.Request, _ *Verifier, _ string) {
			request.Header.Set(TimestampHeader, "1799999600")
		}, wantErr: ErrStale},
		{name: "changed body", mutate: func(_ *httptest.ResponseRecorder, request *http.Request, _ *Verifier, _ string) {
			request.Body = io.NopCloser(strings.NewReader(`{"changed":true}`))
			request.ContentLength = -1
		}, wantErr: ErrInvalid},
		{name: "wrong companion key", mutate: func(_ *httptest.ResponseRecorder, _ *http.Request, _ *Verifier, _ string) {}, wantErr: ErrInvalid},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			verifier := newTestVerifier(t)
			request := signedRequest(t, verifier, "POST", "/api/v1/profile", `{"name":"Ada"}`, "AAAAAAAAAAAAAAAAAAAAAA", time.Unix(1_800_000_000, 0))
			key, err := verifier.ClientKey(testSessionToken)
			if err != nil {
				t.Fatalf("ClientKey(): %v", err)
			}
			test.mutate(httptest.NewRecorder(), request, verifier, key)
			if test.name == "wrong companion key" {
				key = "BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB"
			}
			err = verifier.Verify(request, testSessionToken, key)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("Verify() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}

func TestVerifyRejectsReplayAndCleansExpiredNonces(t *testing.T) {
	t.Parallel()
	verifier := newTestVerifier(t)
	request := signedRequest(t, verifier, "GET", "/api/v1/me", "", "AAAAAAAAAAAAAAAAAAAAAA", time.Unix(1_800_000_000, 0))
	key, _ := verifier.ClientKey(testSessionToken)
	if err := verifier.Verify(request, testSessionToken, key); err != nil {
		t.Fatalf("first Verify(): %v", err)
	}
	request.Body = http.NoBody
	if err := verifier.Verify(request, testSessionToken, key); !errors.Is(err, ErrReplay) {
		t.Fatalf("replayed Verify() error = %v, want %v", err, ErrReplay)
	}

	cache := newReplayCache(1)
	base := time.Unix(100, 0)
	if err := cache.add("first", base, base.Add(time.Second)); err != nil {
		t.Fatalf("cache first add: %v", err)
	}
	if err := cache.add("second", base.Add(2*time.Second), base.Add(3*time.Second)); err != nil {
		t.Fatalf("cache did not clean expired entry: %v", err)
	}
}

func TestVerifyRejectsOldestReplayWindowSecond(t *testing.T) {
	t.Parallel()
	verifier := newTestVerifier(t)
	oldestAcceptedByInclusiveWindow := time.Unix(1_800_000_000, 0).Add(-defaultWindow)
	request := signedRequest(t, verifier, "POST", "/api/v1/purchases", `{}`, "AAAAAAAAAAAAAAAAAAAAAA", oldestAcceptedByInclusiveWindow)
	key, err := verifier.ClientKey(testSessionToken)
	if err != nil {
		t.Fatalf("ClientKey(): %v", err)
	}
	if err := verifier.Verify(request, testSessionToken, key); !errors.Is(err, ErrStale) {
		t.Fatalf("Verify(oldest window second) error = %v, want %v", err, ErrStale)
	}
}

func TestVerifyRejectsOversizedSignedBody(t *testing.T) {
	t.Parallel()
	for _, unknownLength := range []bool{false, true} {
		name := "known length"
		if unknownLength {
			name = "streamed length"
		}
		t.Run(name, func(t *testing.T) {
			verifier := newTestVerifier(t)
			verifier.maxBody = 3
			request := signedRequest(t, verifier, "POST", "/api/v1/example", "four", "AAAAAAAAAAAAAAAAAAAAAA", time.Unix(1_800_000_000, 0))
			if unknownLength {
				request.ContentLength = -1
			}
			key, err := verifier.ClientKey(testSessionToken)
			if err != nil {
				t.Fatalf("ClientKey(): %v", err)
			}
			if err := verifier.Verify(request, testSessionToken, key); !errors.Is(err, ErrBodyTooLarge) {
				t.Fatalf("Verify(oversized body) error = %v, want %v", err, ErrBodyTooLarge)
			}
		})
	}
}

func TestVerifyFailsClosedWhenReplayCacheIsFull(t *testing.T) {
	t.Parallel()
	verifier := newTestVerifier(t)
	verifier.replays = newReplayCache(1)
	key, err := verifier.ClientKey(testSessionToken)
	if err != nil {
		t.Fatalf("ClientKey(): %v", err)
	}
	first := signedRequest(t, verifier, "GET", "/api/v1/me", "", "AAAAAAAAAAAAAAAAAAAAAA", time.Unix(1_800_000_000, 0))
	if err := verifier.Verify(first, testSessionToken, key); err != nil {
		t.Fatalf("first Verify(): %v", err)
	}
	second := signedRequest(t, verifier, "GET", "/api/v1/me", "", "BBBBBBBBBBBBBBBBBBBBBB", time.Unix(1_800_000_000, 0))
	if err := verifier.Verify(second, testSessionToken, key); !errors.Is(err, ErrReplay) {
		t.Fatalf("capacity Verify() error = %v, want wrapped %v", err, ErrReplay)
	}
}

func TestReplayCacheSupportsConcurrentSessions(t *testing.T) {
	t.Parallel()
	cache := newReplayCache(128)
	base := time.Unix(100, 0)
	errorsChannel := make(chan error, 64)
	var workers sync.WaitGroup
	for index := range 64 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			errorsChannel <- cache.add(strconv.Itoa(index), base, base.Add(time.Minute))
		}()
	}
	workers.Wait()
	close(errorsChannel)
	for err := range errorsChannel {
		if err != nil {
			t.Fatalf("concurrent cache add: %v", err)
		}
	}
}

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
	t.Helper()
	request := httptest.NewRequest(method, target, strings.NewReader(body))
	key, err := verifier.ClientKey(testSessionToken)
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
