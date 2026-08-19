package billing

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
	jobpayload "github.com/txyyddss/Remna-User-Panel/internal/outbox"
)

// PaymentAnnouncementChatIDSetting selects the standalone Telegram channel
// that receives successful provider-payment announcements.
const PaymentAnnouncementChatIDSetting = "telegram.payment_announcement_chat_id"

// PaymentAnnouncementSettings loads the optional standalone channel ID.
type PaymentAnnouncementSettings interface {
	Optional(context.Context, string) (string, error)
}

// PaymentAnnouncementSender is the narrow Telegram delivery contract.
type PaymentAnnouncementSender interface {
	SendMarkdownV2Message(context.Context, int64, int64, string) error
}

// PaymentAnnouncementWorker sends immutable settlement snapshots from the
// durable outbox.
type PaymentAnnouncementWorker struct {
	settings PaymentAnnouncementSettings
	sender   PaymentAnnouncementSender
}

// NewPaymentAnnouncementWorker creates the durable Telegram delivery handler.
func NewPaymentAnnouncementWorker(settings PaymentAnnouncementSettings, sender PaymentAnnouncementSender) *PaymentAnnouncementWorker {
	return &PaymentAnnouncementWorker{settings: settings, sender: sender}
}

// HandleOutbox sends one payment announcement or skips delivery when no
// standalone channel has been configured.
func (w *PaymentAnnouncementWorker) HandleOutbox(ctx context.Context, job model.OutboxJob) error {
	payload, err := jobpayload.DecodePaymentSuccessAnnouncement(job)
	if err != nil {
		return err
	}
	if err := validatePaymentSuccessAnnouncement(payload); err != nil {
		return err
	}
	rawChatID, err := w.settings.Optional(ctx, PaymentAnnouncementChatIDSetting)
	if err != nil {
		return fmt.Errorf("load payment announcement channel: %w", err)
	}
	rawChatID = strings.TrimSpace(rawChatID)
	if rawChatID == "" {
		return nil
	}
	chatID, err := strconv.ParseInt(rawChatID, 10, 64)
	if err != nil || chatID == 0 {
		return errors.New("payment announcement channel must be a non-zero integer")
	}
	if err := w.sender.SendMarkdownV2Message(ctx, chatID, 0, formatPaymentSuccessAnnouncement(payload)); err != nil {
		return fmt.Errorf("send payment success announcement: %w", err)
	}
	return nil
}

func formatPaymentSuccessAnnouncement(payload jobpayload.PaymentSuccessAnnouncement) string {
	return fmt.Sprintf("*Payment successful*\n*Provider:* %s\n*Channel:* %s\n*TXB amount:* %s\n*Paid amount:* %s %s\n*Username:* %s",
		markdownV2Escape(paymentAnnouncementProvider(payload.Provider)), markdownV2Escape(paymentAnnouncementChannel(payload.Provider, payload.Channel)),
		markdownV2Escape(model.TXBMoney(payload.TXBMinor).Display), markdownV2Escape(payload.PayableAmount),
		markdownV2Escape(payload.PayableCurrency), markdownV2Escape(payload.Username))
}

var markdownV2Escaper = strings.NewReplacer(
	"\\", "\\\\", "_", "\\_", "*", "\\*", "[", "\\[", "]", "\\]", "(", "\\(", ")", "\\)",
	"~", "\\~", "`", "\\`", ">", "\\>", "#", "\\#", "+", "\\+", "-", "\\-", "=", "\\=", "|", "\\|", "{", "\\{", "}", "\\}", ".", "\\.", "!", "\\!",
)

func markdownV2Escape(value string) string {
	return markdownV2Escaper.Replace(value)
}

func validatePaymentSuccessAnnouncement(payload jobpayload.PaymentSuccessAnnouncement) error {
	amount, err := ParseDecimal(payload.PayableAmount)
	if err != nil || !amount.Positive() {
		return errors.New("payment announcement payable amount is invalid")
	}
	expectedCurrency := strings.ToUpper(currencyCode(payload.Provider))
	if expectedCurrency == "" || payload.PayableCurrency != expectedCurrency {
		return errors.New("payment announcement provider currency is invalid")
	}
	if payload.Provider != "ezpay" && payload.Provider != "bepusdt" && payload.Provider != "stars" {
		return errors.New("payment announcement provider is unsupported")
	}
	return nil
}

func paymentAnnouncementProvider(provider string) string {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "ezpay":
		return "EZPay"
	case "bepusdt":
		return "BEPUSDT"
	case "stars":
		return "Telegram Stars"
	default:
		return strings.ToUpper(strings.TrimSpace(provider))
	}
}

func paymentAnnouncementChannel(provider, channel string) string {
	channel = strings.TrimSpace(channel)
	if channel == "" {
		_, _, legacyRail, err := ParseMethodSelection(provider)
		if err == nil {
			channel = legacyRail
		}
	}
	if name := methodName(provider, channel); name != "" {
		return name
	}
	if channel != "" {
		return channel
	}
	return "Default"
}
