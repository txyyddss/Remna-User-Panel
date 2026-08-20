package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAuthenticationLimiterIsPerClientAndRefills(t *testing.T) {
	limiter := newAuthLimiter()
	now := time.Unix(100, 0)
	limiter.now = func() time.Time { return now }
	server := &Server{authLimiter: limiter}
	handler := server.limitAuthentication(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	for attempt := 0; attempt < authLimitCapacity; attempt++ {
		request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/telegram", nil)
		request.RemoteAddr = "192.0.2.1"
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusNoContent {
			t.Fatalf("attempt %d status = %d", attempt, response.Code)
		}
	}
	blocked := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/telegram", nil)
	request.RemoteAddr = "192.0.2.1"
	handler.ServeHTTP(blocked, request)
	if blocked.Code != http.StatusTooManyRequests || blocked.Header().Get("Retry-After") != "10" {
		t.Fatalf("blocked response = %d, retry %q", blocked.Code, blocked.Header().Get("Retry-After"))
	}

	other := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/telegram", nil)
	request.RemoteAddr = "192.0.2.2"
	handler.ServeHTTP(other, request)
	if other.Code != http.StatusNoContent {
		t.Fatalf("other client status = %d", other.Code)
	}
	now = now.Add(authLimitRefill)
	refilled := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/telegram", nil)
	request.RemoteAddr = "192.0.2.1"
	handler.ServeHTTP(refilled, request)
	if refilled.Code != http.StatusNoContent {
		t.Fatalf("refilled status = %d", refilled.Code)
	}
}
