package admin

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/txyyddss/Remna-User-Panel/internal/iplookup"
	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/database"
)

type ipSettingsRepo struct {
	*adminSettingsRepository
	quotas map[string]*int
}

func (r *ipSettingsRepo) ListCombos(context.Context, bool) ([]model.Combo, error) {
	return []model.Combo{{ID: "combo", IPLookupQuota: r.quotas["combo"]}}, nil
}
func (r *ipSettingsRepo) SaveIPLookupSettings(ctx context.Context, actor string, c iplookup.Config, secrets map[string]string, quotas map[string]*int) error {
	raw, err := json.Marshal(c)
	if err != nil {
		return err
	}
	if err := r.PutSetting(ctx, iplookup.SettingKey, string(raw), false, &actor); err != nil {
		return err
	}
	for id, v := range secrets {
		if err := r.PutSetting(ctx, iplookup.CredentialKey(id), v, true, &actor); err != nil {
			return err
		}
	}
	r.quotas = quotas
	return nil
}

func TestIPLookupCredentialsAreEncryptedAndWriteOnly(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := &ipSettingsRepo{adminSettingsRepository: newAdminSettingsRepository()}
	s := NewSettingsService(repo, testVault(t))
	q := 1
	input := iplookup.AdminSettings{Config: iplookup.DefaultConfig(), Credentials: map[string]iplookup.CredentialEdit{"ipapi": {Value: "ip-lookup-test-secret"}}, ComboQuotas: map[string]*int{"combo": &q}}
	if err := s.PutIPLookupSettings(ctx, "admin", input); err != nil {
		t.Fatal(err)
	}
	stored := repo.settings[iplookup.CredentialKey("ipapi")]
	if !stored.Encrypted || stored.Value == "ip-lookup-test-secret" {
		t.Fatal("credential stored as plaintext")
	}
	view, err := s.IPLookupSettings(ctx)
	if err != nil || !view.Configured["ipapi"] || len(view.Credentials) != 0 {
		t.Fatalf("unsafe view=%+v err=%v", view, err)
	}
	input.Credentials["ipapi"] = iplookup.CredentialEdit{}
	if err := s.PutIPLookupSettings(ctx, "admin", input); err != nil {
		t.Fatal(err)
	}
	if repo.settings[iplookup.CredentialKey("ipapi")].Value != stored.Value {
		t.Fatal("blank credential replaced secret")
	}
	input.Credentials["ipapi"] = iplookup.CredentialEdit{Clear: true}
	if err := s.PutIPLookupSettings(ctx, "admin", input); err != nil {
		t.Fatal(err)
	}
	if repo.settings[iplookup.CredentialKey("ipapi")].Value != "" {
		t.Fatal("explicit clear failed")
	}
	if err := s.Put(ctx, "admin", iplookup.SettingKey, `{"enabled":true}`); err == nil {
		t.Fatal("generic setting route bypassed atomic validation")
	}
	if _, err := s.Plaintext(ctx, "missing-setting"); err != database.ErrNotFound {
		t.Fatalf("missing=%v", err)
	}
}
