package app

import (
	"testing"

	"github.com/txyyddss/Remna-User-Panel/internal/integrations/telegram"
)

func TestNormalizeStarTransactionDirections(t *testing.T) {
	t.Parallel()

	user := telegram.User{ID: 42, FirstName: "Ada"}
	partner := &telegram.TransactionPartnerUser{Type: "user", TransactionType: "invoice_payment", User: user, InvoicePayload: "order-1"}
	tests := []struct {
		name       string
		input      telegram.StarTransaction
		wantOK     bool
		wantRefund bool
	}{
		{name: "incoming payment", input: telegram.StarTransaction{ID: "charge-1", Amount: 25, Source: partner}, wantOK: true},
		{name: "outgoing refund", input: telegram.StarTransaction{ID: "charge-1", Amount: -25, Receiver: partner}, wantOK: true, wantRefund: true},
		{name: "fractional transaction", input: telegram.StarTransaction{ID: "charge-1", Amount: 25, NanostarAmount: 1, Source: partner}},
		{name: "unrelated partner", input: telegram.StarTransaction{ID: "charge-1", Amount: 25, Source: &telegram.TransactionPartnerUser{Type: "fragment"}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			event, refund, ok := normalizeStarTransaction(test.input)
			if ok != test.wantOK || refund != test.wantRefund {
				t.Fatalf("normalizeStarTransaction() = (%+v, %t, %t), want ok=%t refund=%t", event, refund, ok, test.wantOK, test.wantRefund)
			}
			if test.wantOK && (event.OrderID != "order-1" || event.TradeID != "charge-1" || event.PayableAmount != "25" || event.TelegramID == nil || *event.TelegramID != 42) {
				t.Fatalf("normalized event = %+v", event)
			}
			if test.wantOK && !test.wantRefund && event.DedupeKey != "charge-1" {
				t.Fatalf("incoming dedupe key = %q", event.DedupeKey)
			}
		})
	}
}
