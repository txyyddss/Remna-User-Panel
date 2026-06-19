package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"remna-user-panel/internal/config"
	"remna-user-panel/internal/payments"
)

func TestHealthz(t *testing.T) {
	router := BackendRouter(config.Settings{PanelWebhookPath: "/webhook/panel"}, nil, nil, payments.NewRegistry(config.Settings{}, nil), nil)
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}
}
