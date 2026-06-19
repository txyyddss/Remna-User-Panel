package remnawave

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"remna-user-panel/internal/config"
	appsettings "remna-user-panel/internal/settings"
)

func TestClientAddsAPIBaseAndBearerToken(t *testing.T) {
	t.Parallel()

	requestPath := ""
	authHeader := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestPath = r.URL.Path
		authHeader = r.Header.Get("Authorization")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"response": map[string]any{"ok": true},
		})
	}))
	defer server.Close()

	client := NewClient(config.Settings{
		PanelAPIURL:          server.URL,
		PanelAPIKey:          "secret",
		PanelAPITotalTimeout: time.Second,
	}, appsettings.NewStore(nil))

	if err := client.ResetUserTraffic(context.Background(), "user-uuid"); err != nil {
		t.Fatalf("ResetUserTraffic() error = %v", err)
	}
	if requestPath != "/api/users/user-uuid/actions/reset-traffic" {
		t.Fatalf("path = %q, want /api/users/user-uuid/actions/reset-traffic", requestPath)
	}
	if authHeader != "Bearer secret" {
		t.Fatalf("Authorization = %q, want Bearer secret", authHeader)
	}
}

func TestGetSubscriptionPageConfigsDecodesEnvelope(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/subscription-page-configs" {
			t.Fatalf("path = %q, want /api/subscription-page-configs", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"response": map[string]any{
				"items": []map[string]any{
					{"uuid": "cfg-1", "name": "Default"},
				},
			},
		})
	}))
	defer server.Close()

	client := NewClient(config.Settings{
		PanelAPIURL:          server.URL + "/api",
		PanelAPIKey:          "secret",
		PanelAPITotalTimeout: time.Second,
	}, appsettings.NewStore(nil))

	configs, err := client.GetSubscriptionPageConfigs(context.Background())
	if err != nil {
		t.Fatalf("GetSubscriptionPageConfigs() error = %v", err)
	}
	if len(configs) != 1 || configs[0]["uuid"] != "cfg-1" {
		t.Fatalf("configs = %#v, want cfg-1", configs)
	}
}
