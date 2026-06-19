package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"remna-user-panel/internal/config"
	"remna-user-panel/internal/i18n"
	"remna-user-panel/internal/payments"
	"remna-user-panel/internal/webassets"
)

func TestBootstrapDefaultsToChinese(t *testing.T) {
	catalog := &i18n.Catalog{}
	router := WebAppRouter(
		config.Settings{DefaultLanguage: "zh"},
		nil,
		catalog,
		webassets.Paths{TemplatesDir: "../../internal/webassets/templates", ThemesDir: "../../internal/webassets/themes"},
		payments.NewRegistry(config.Settings{}, nil),
		nil,
	)
	request := httptest.NewRequest(http.MethodGet, "/api/bootstrap", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}
	if !strings.Contains(response.Body.String(), `"language":"zh"`) {
		t.Fatalf("bootstrap response does not include zh language: %s", response.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode bootstrap response: %v", err)
	}
	if payload["ok"] != true || payload["i18n"] == nil || payload["messages"] == nil {
		t.Fatalf("bootstrap response missing i18n compatibility fields: %#v", payload)
	}
}
