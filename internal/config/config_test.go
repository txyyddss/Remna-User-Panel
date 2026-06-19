package config

import "testing"

func TestLoadDefaultsToChinese(t *testing.T) {
	t.Setenv("POSTGRES_USER", "shop")
	t.Setenv("POSTGRES_PASSWORD", "secret")
	t.Setenv("DEFAULT_LANGUAGE", "")

	settings, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if settings.DefaultLanguage != "zh" {
		t.Fatalf("DefaultLanguage = %q, want zh", settings.DefaultLanguage)
	}
}

func TestSubscriptionWebhookPathUsesToken(t *testing.T) {
	settings := Settings{BotToken: "123:abc"}
	if got, want := settings.WebhookPath(), "/webhook/123:abc"; got != want {
		t.Fatalf("WebhookPath() = %q, want %q", got, want)
	}
}
