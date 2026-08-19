package billing

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
	jobpayload "github.com/txyyddss/Remna-User-Panel/internal/outbox"
)

func TestPaymentAnnouncementFormatting(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		payload  jobpayload.PaymentSuccessAnnouncement
		contains []string
	}{
		{name: "EZPay CNY", payload: paymentAnnouncementFixture("ezpay", "alipay", "12.34", "CNY"),
			contains: []string{"*Provider:* EZPay", "*Channel:* Alipay", "*TXB amount:* 25\\.00 TXB", "*Paid amount:* 12\\.34 CNY", "*Username:* @ada"}},
		{name: "BEPUSDT USD", payload: paymentAnnouncementFixture("bepusdt", "usdt.trc20", "3.21", "USD"),
			contains: []string{"*Provider:* BEPUSDT", "*Channel:* USDT TRC20", "*Paid amount:* 3\\.21 USD"}},
		{name: "legacy EZPay channel", payload: paymentAnnouncementFixture("ezpay", "", "12.34", "CNY"),
			contains: []string{"*Provider:* EZPay", "*Channel:* Alipay"}},
		{name: "Stars fallback", payload: paymentAnnouncementFixture("stars", "", "7", "XTR"),
			contains: []string{"*Provider:* Telegram Stars", "*Channel:* Telegram Stars", "*Paid amount:* 7 XTR"}},
		{name: "MarkdownV2 dynamic text", payload: paymentAnnouncementFixture("ezpay", "fee_test", "12.34", "CNY"),
			contains: []string{"fee\\_test", "@ada"}},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			message := formatPaymentSuccessAnnouncement(test.payload)
			for _, fragment := range test.contains {
				if !strings.Contains(message, fragment) {
					t.Errorf("announcement %q does not contain %q", message, fragment)
				}
			}
		})
	}
}

func TestPaymentAnnouncementWorkerSkipRetryAndDelivery(t *testing.T) {
	t.Parallel()

	sendFailure := errors.New("telegram temporarily unavailable")
	tests := []struct {
		name       string
		setting    string
		settingErr error
		sendErr    error
		wantErr    bool
		wantCalls  int
	}{
		{name: "missing channel skips"},
		{name: "settings failure retries", settingErr: errors.New("settings unavailable"), wantErr: true},
		{name: "invalid channel retries", setting: "channel", wantErr: true},
		{name: "Telegram failure retries", setting: "-100123", sendErr: sendFailure, wantErr: true, wantCalls: 1},
		{name: "success", setting: "-100123", wantCalls: 1},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			settings := &paymentAnnouncementSettingsStub{value: test.setting, err: test.settingErr}
			sender := &paymentAnnouncementSenderStub{err: test.sendErr}
			worker := NewPaymentAnnouncementWorker(settings, sender)
			err := worker.HandleOutbox(context.Background(), paymentAnnouncementJob(t, paymentAnnouncementFixture("ezpay", "wxpay", "12.34", "CNY")))
			if (err != nil) != test.wantErr {
				t.Fatalf("HandleOutbox() error = %v, want error %t", err, test.wantErr)
			}
			if sender.calls != test.wantCalls {
				t.Fatalf("SendMarkdownV2Message() calls = %d, want %d", sender.calls, test.wantCalls)
			}
			if test.wantCalls == 1 && sender.chatID != -100123 {
				t.Fatalf("SendMarkdownV2Message() chat ID = %d", sender.chatID)
			}
			if test.sendErr != nil && !errors.Is(err, sendFailure) {
				t.Fatalf("HandleOutbox() error = %v, want wrapped send failure", err)
			}
		})
	}
}

func paymentAnnouncementFixture(provider, channel, amount, currency string) jobpayload.PaymentSuccessAnnouncement {
	return jobpayload.PaymentSuccessAnnouncement{
		OrderID: "order-1", Provider: provider, Channel: channel, TXBMinor: 2_500,
		PayableAmount: amount, PayableCurrency: currency, Username: "@ada",
	}
}

func paymentAnnouncementJob(t *testing.T, payload jobpayload.PaymentSuccessAnnouncement) model.OutboxJob {
	t.Helper()
	encoded, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json.Marshal(): %v", err)
	}
	return model.OutboxJob{Kind: jobpayload.PaymentSuccessAnnouncementKind, Payload: string(encoded)}
}

type paymentAnnouncementSettingsStub struct {
	value string
	err   error
}

func (s *paymentAnnouncementSettingsStub) Optional(context.Context, string) (string, error) {
	return s.value, s.err
}

type paymentAnnouncementSenderStub struct {
	chatID int64
	calls  int
	err    error
}

func (s *paymentAnnouncementSenderStub) SendMarkdownV2Message(_ context.Context, chatID, _ int64, _ string) error {
	s.calls++
	s.chatID = chatID
	return s.err
}
