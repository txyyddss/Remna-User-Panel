package database

import (
	"context"
	"testing"
)

func TestSetDisplayCurrencyPersistsOnlyTheTargetUser(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)
	first := createTestUser(t, store, 70101)
	second := createTestUser(t, store, 70102)

	updated, err := store.SetDisplayCurrency(context.Background(), first.ID, "CNY")
	if err != nil || updated.DisplayCurrency != "CNY" {
		t.Fatalf("SetDisplayCurrency() = (%+v, %v)", updated, err)
	}
	untouched, err := store.UserByID(context.Background(), second.ID)
	if err != nil || untouched.DisplayCurrency != "TXB" {
		t.Fatalf("unrelated user = (%+v, %v)", untouched, err)
	}
}
