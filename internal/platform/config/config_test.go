package config

import (
	"encoding/base64"
	"testing"
)

func TestLoadValidatesCanonicalPublicOrigin(t *testing.T) {
	setValidEnvironment(t)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load(): %v", err)
	}
	if cfg.PublicBaseURL.String() != "https://carpool.example" || cfg.AdminTelegramID != 42 || len(cfg.MasterKey) != 32 {
		t.Fatalf("config = %+v", cfg)
	}

	tests := []struct {
		name      string
		publicURL string
		insecure  string
		wantError bool
	}{
		{name: "local HTTP opt in", publicURL: "http://127.0.0.1:8080", insecure: "true"},
		{name: "HTTP rejected", publicURL: "http://carpool.example", wantError: true},
		{name: "non HTTP scheme rejected", publicURL: "ftp://carpool.example", insecure: "true", wantError: true},
		{name: "path rejected", publicURL: "https://carpool.example/app", wantError: true},
		{name: "credentials rejected", publicURL: "https://user@carpool.example", wantError: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			setValidEnvironment(t)
			t.Setenv("PUBLIC_BASE_URL", test.publicURL)
			t.Setenv("ALLOW_INSECURE_HTTP", test.insecure)
			_, err := Load()
			if test.wantError && err == nil {
				t.Fatal("Load() unexpectedly succeeded")
			}
			if !test.wantError && err != nil {
				t.Fatalf("Load(): %v", err)
			}
		})
	}
}

func setValidEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("ADMIN_TELEGRAM_ID", "42")
	t.Setenv("TELEGRAM_BOT_TOKEN", "42:test-token")
	t.Setenv("PUBLIC_BASE_URL", "https://carpool.example")
	t.Setenv("CONFIG_MASTER_KEY", base64.StdEncoding.EncodeToString(make([]byte, 32)))
	t.Setenv("TZ", "UTC")
	t.Setenv("ALLOW_INSECURE_HTTP", "")
}
