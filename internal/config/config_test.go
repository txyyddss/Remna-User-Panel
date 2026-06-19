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

func TestDatabaseURLPrecedence(t *testing.T) {
	// SHOP_DATABASE_URL 优先于 DATABASE_URL
	t.Setenv("SHOP_DATABASE_URL", "postgres://shop_user:shop_pass@shop_host:5432/shop_db?sslmode=disable")
	t.Setenv("DATABASE_URL", "postgres://other_user:other_pass@other_host:5432/other_db")
	t.Setenv("POSTGRES_USER", "unused")
	t.Setenv("POSTGRES_PASSWORD", "unused")

	settings, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if settings.DatabaseURL != "postgres://shop_user:shop_pass@shop_host:5432/shop_db?sslmode=disable" {
		t.Fatalf("DatabaseURL = %q, want SHOP_DATABASE_URL value", settings.DatabaseURL)
	}
}

func TestDatabaseURLFallback(t *testing.T) {
	// DATABASE_URL 作为回退
	t.Setenv("DATABASE_URL", "postgres://fb_user:fb_pass@fb_host:5432/fb_db")
	t.Setenv("POSTGRES_USER", "unused")
	t.Setenv("POSTGRES_PASSWORD", "unused")

	settings, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if settings.DatabaseURL != "postgres://fb_user:fb_pass@fb_host:5432/fb_db" {
		t.Fatalf("DatabaseURL = %q, want DATABASE_URL value", settings.DatabaseURL)
	}
}

func TestRedisURLFromComponents(t *testing.T) {
	// REDIS_HOST + REDIS_PORT 构建 REDIS_URL
	t.Setenv("POSTGRES_USER", "u")
	t.Setenv("POSTGRES_PASSWORD", "p")
	t.Setenv("REDIS_HOST", "myredis")
	t.Setenv("REDIS_PORT", "6380")

	settings, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if settings.RedisURL != "redis://myredis:6380/0" {
		t.Fatalf("RedisURL = %q, want redis://myredis:6380/0", settings.RedisURL)
	}
}

func TestRedisURLPrecedence(t *testing.T) {
	// REDIS_URL 优先于 REDIS_HOST/REDIS_PORT
	t.Setenv("POSTGRES_USER", "u")
	t.Setenv("POSTGRES_PASSWORD", "p")
	t.Setenv("REDIS_URL", "redis://explicit:6379/1")
	t.Setenv("REDIS_HOST", "ignored")
	t.Setenv("REDIS_PORT", "1234")

	settings, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if settings.RedisURL != "redis://explicit:6379/1" {
		t.Fatalf("RedisURL = %q, want REDIS_URL value", settings.RedisURL)
	}
}
