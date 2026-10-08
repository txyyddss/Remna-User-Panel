package admin

import (
	"context"
	"errors"
	"testing"
)

func TestPMSettingsValidateForumBeforeEnableAndDestinationChange(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repository := newAdminSettingsRepository()
	service := NewSettingsService(repository, testVault(t))
	if err := service.Put(ctx, "admin", "telegram.pm.enabled", "true"); err == nil {
		t.Fatal("enabled without destination")
	}
	if err := service.Put(ctx, "admin", "telegram.pm.group_chat_id", "-123"); err == nil {
		t.Fatal("basic group accepted")
	}
	if err := service.Put(ctx, "admin", "telegram.pm.group_chat_id", "-100123"); err != nil {
		t.Fatal(err)
	}
	if err := service.Put(ctx, "admin", "telegram.pm.enabled", "true"); err == nil {
		t.Fatal("enabled without queued validator")
	}
	var validated []int64
	service.SetPMForumValidator(func(_ context.Context, id int64) error {
		validated = append(validated, id)
		if id == -100124 {
			return errors.New("not a forum")
		}
		return nil
	})
	if err := service.Put(ctx, "admin", "telegram.pm.enabled", "true"); err != nil {
		t.Fatal(err)
	}
	if err := service.Put(ctx, "admin", "telegram.pm.group_chat_id", "-100124"); err == nil {
		t.Fatal("unusable destination replaced active forum")
	}
	if group, err := service.Optional(ctx, "telegram.pm.group_chat_id"); err != nil || group != "-100123" {
		t.Fatalf("destination=%s,%v", group, err)
	}
	if err := service.Put(ctx, "admin", "telegram.pm.group_chat_id", ""); err == nil {
		t.Fatal("cleared enabled destination")
	}
	if len(validated) != 2 {
		t.Fatalf("queued validations=%v", validated)
	}
	if err := service.Put(ctx, "admin", "telegram.pm.enabled", "false"); err != nil {
		t.Fatal(err)
	}
	if err := service.Put(ctx, "admin", "telegram.pm.group_chat_id", ""); err != nil {
		t.Fatal(err)
	}
}
