package connectivity

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestDecodeConfigBoundariesAndSecrets(t *testing.T) {
	for _, test := range []struct {
		name  string
		value string
		valid bool
	}{
		{"defaults", "", true},
		{"configured", `{"remnawaveUserId":42,"scheduledEnabled":true}`, true},
		{"minimum", `{"remnawaveUserId":42,"intervalSeconds":60,"timeoutSeconds":1}`, true},
		{"maximum", `{"remnawaveUserId":42,"intervalSeconds":86400,"timeoutSeconds":60}`, true},
		{"missing_account", `{"scheduledEnabled":true}`, false},
		{"negative_account", `{"remnawaveUserId":-1}`, false},
		{"short_interval", `{"intervalSeconds":59}`, false},
		{"long_interval", `{"intervalSeconds":86401}`, false},
		{"zero_timeout", `{"timeoutSeconds":0}`, false},
		{"long_timeout", `{"timeoutSeconds":61}`, false},
		{"http", `{"probeUrl":"http://example.com/"}`, false},
		{"credentials", `{"probeUrl":"https://secret:password@example.com/"}`, false},
		{"relative", `{"probeUrl":"/generate_204"}`, false},
		{"unknown", `{"password":"secret"}`, false},
		{"null", `null`, false},
		{"trailing_object", `{} {}`, false},
		{"wrong_type", `{"remnawaveUserId":"42"}`, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			cfg, err := DecodeConfig(test.value)
			if (err == nil) != test.valid {
				t.Fatalf("valid=%t, config=%+v error=%v", test.valid, cfg, err)
			}
			if err != nil && strings.Contains(err.Error(), "secret") {
				t.Fatal("configuration error exposed a credential")
			}
		})
	}
	cfg := DefaultConfig()
	if cfg.ScheduledEnabled || cfg.RemnawaveUserID != 0 || cfg.IntervalSeconds != 300 || cfg.TimeoutSeconds != 15 ||
		cfg.ProbeURL != "https://cp.cloudflare.com/generate_204" {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
}

func TestConfigHashSeparatesChangedProbeIdentity(t *testing.T) {
	cfg := DefaultConfig()
	cfg.RemnawaveUserID = 42
	first := ConfigHash(cfg)
	if len(first) != 64 || ConfigHash(cfg) != first {
		t.Fatal("configuration fingerprint is unstable")
	}
	for _, change := range []func(*Config){
		func(c *Config) { c.RemnawaveUserID++ },
		func(c *Config) { c.ProbeURL = "https://example.com/generate_204" },
		func(c *Config) { c.TimeoutSeconds++ },
		func(c *Config) { c.IntervalSeconds++ },
		func(c *Config) { c.ScheduledEnabled = true },
	} {
		copy := cfg
		change(&copy)
		if ConfigHash(copy) == first {
			t.Fatal("changed settings reused old latest-result identity")
		}
	}
}

func TestErrorCodeNeverExposesProviderSecrets(t *testing.T) {
	for _, err := range []error{
		errors.New("https://token:password@example.com/private"),
		&CodeError{Code: "secret credential"},
		&CodeError{Code: "CONNECTIVITY_ACCOUNT_UNAVAILABLE", Err: errors.New("secret token")},
	} {
		code := ErrorCode(err)
		if !safeErrorCode.MatchString(code) || strings.Contains(code, "secret") {
			t.Fatalf("unsafe public error code: %q", code)
		}
	}
	if ErrorCode(context.Canceled) != "CONNECTIVITY_INTERRUPTED" || ErrorCode(context.DeadlineExceeded) != "CONNECTIVITY_PROBE_TIMEOUT" {
		t.Fatal("cancellation classification was lost")
	}
}
