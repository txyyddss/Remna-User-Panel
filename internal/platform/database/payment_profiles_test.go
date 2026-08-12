package database

import (
	"context"
	"testing"
)

func TestPaymentProfilesAllowIndependentProviderAccounts(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t)
	first, err := store.SavePaymentProfile(ctx, PaymentProfileInput{
		ID: "ezpay-main", Provider: "ezpay", ProviderName: "Main EZPay", EnabledChannels: []string{"alipay"},
		Endpoint: "https://ezpay-main.example", MerchantID: "main", CredentialCiphertext: "main-secret", Enabled: true,
	})
	if err != nil {
		t.Fatalf("SavePaymentProfile(first): %v", err)
	}
	second, err := store.SavePaymentProfile(ctx, PaymentProfileInput{
		ID: "ezpay-backup", Provider: "ezpay", ProviderName: "Backup EZPay", EnabledChannels: []string{"wxpay"},
		Endpoint: "https://ezpay-backup.example", MerchantID: "backup", CredentialCiphertext: "backup-secret", Enabled: true,
	})
	if err != nil {
		t.Fatalf("SavePaymentProfile(second): %v", err)
	}
	if first.ID != "ezpay-main" || second.ID != "ezpay-backup" {
		t.Fatalf("saved profile IDs = %q, %q", first.ID, second.ID)
	}

	profiles, err := store.ListPaymentProfiles(ctx)
	if err != nil {
		t.Fatalf("ListPaymentProfiles(): %v", err)
	}
	found := map[string]bool{}
	for _, profile := range profiles {
		if profile.ID == first.ID || profile.ID == second.ID {
			found[profile.ID] = true
		}
	}
	if len(found) != 2 {
		t.Fatalf("ListPaymentProfiles() = %+v", profiles)
	}
	backup, err := store.PaymentProfileRecordByID(ctx, second.ID, "wxpay")
	if err != nil || backup.ProviderName != "Backup EZPay" || len(backup.EnabledChannels) != 1 || backup.EnabledChannels[0] != "wxpay" {
		t.Fatalf("PaymentProfileRecordByID() = (%+v, %v)", backup, err)
	}
}
