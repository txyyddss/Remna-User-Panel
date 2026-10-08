package admin

import (
	"context"
	"testing"
)

func TestTurnstileRequiresEncryptedKeysBeforeEnable(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repository := newAdminSettingsRepository()
	service := NewSettingsService(repository, testVault(t))
	if err := service.Put(ctx, "admin", "captcha.turnstile.enabled", "true"); err == nil {
		t.Fatal("enabled without keys")
	}
	if err := service.Put(ctx, "admin", "captcha.turnstile.site_key", "site-key"); err != nil {
		t.Fatal(err)
	}
	if err := service.Put(ctx, "admin", "captcha.turnstile.enabled", "true"); err == nil {
		t.Fatal("enabled without secret")
	}
	if err := service.Put(ctx, "admin", "captcha.turnstile.secret_key", "private-secret"); err != nil {
		t.Fatal(err)
	}
	stored := repository.settings["captcha.turnstile.secret_key"]
	if !stored.Encrypted || stored.Value == "private-secret" {
		t.Fatal("unencrypted secret")
	}
	if err := service.Put(ctx, "admin", "captcha.turnstile.enabled", "true"); err != nil {
		t.Fatal(err)
	}
	if err := service.Put(ctx, "admin", "captcha.turnstile.site_key", ""); err == nil {
		t.Fatal("cleared enabled site key")
	}
	if err := service.Put(ctx, "admin", "captcha.turnstile.enabled", "false"); err != nil {
		t.Fatal(err)
	}
	if err := service.Put(ctx, "admin", "captcha.turnstile.site_key", ""); err != nil {
		t.Fatal(err)
	}
}
