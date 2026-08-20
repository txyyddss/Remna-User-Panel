package outbox

import (
	"testing"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
)

func TestDecodePaymentSuccessAnnouncementProviderNameCompatibility(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		json string
		want string
	}{
		{name: "legacy payload", json: `{"orderId":"order-1","provider":"ezpay","channel":"alipay","txbMinor":100,"payableAmount":"1.00","payableCurrency":"CNY","username":"@ada"}`},
		{name: "provider snapshot", json: `{"orderId":"order-1","provider":"ezpay","providerName":"  Main EZPay  ","channel":"alipay","txbMinor":100,"payableAmount":"1.00","payableCurrency":"CNY","username":"@ada"}`, want: "Main EZPay"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			payload, err := DecodePaymentSuccessAnnouncement(model.OutboxJob{Kind: PaymentSuccessAnnouncementKind, Payload: test.json})
			if err != nil {
				t.Fatalf("DecodePaymentSuccessAnnouncement(): %v", err)
			}
			if payload.ProviderName != test.want {
				t.Errorf("ProviderName = %q, want %q", payload.ProviderName, test.want)
			}
		})
	}
}
