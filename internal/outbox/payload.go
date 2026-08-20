package outbox

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
)

// PaymentSuccessAnnouncement is the immutable payment snapshot carried by the
// durable Telegram announcement job.
type PaymentSuccessAnnouncement struct {
	OrderID         string `json:"orderId"`
	Provider        string `json:"provider"`
	ProviderName    string `json:"providerName,omitempty"`
	Channel         string `json:"channel"`
	TXBMinor        int64  `json:"txbMinor"`
	PayableAmount   string `json:"payableAmount"`
	PayableCurrency string `json:"payableCurrency"`
	Username        string `json:"username"`
}

// TargetID extracts one required identifier from a canonical typed job payload.
func TargetID(job model.OutboxJob, field string) (string, error) {
	var payload map[string]string
	if err := json.Unmarshal([]byte(job.Payload), &payload); err != nil {
		return "", fmt.Errorf("decode %s job payload: %w", job.Kind, err)
	}
	value := strings.TrimSpace(payload[field])
	if value == "" {
		return "", errors.New("outbox payload is missing " + field)
	}
	return value, nil
}

// DecodePaymentSuccessAnnouncement validates one canonical announcement job.
func DecodePaymentSuccessAnnouncement(job model.OutboxJob) (PaymentSuccessAnnouncement, error) {
	if job.Kind != PaymentSuccessAnnouncementKind {
		return PaymentSuccessAnnouncement{}, errors.New("unexpected payment announcement job kind")
	}
	var payload PaymentSuccessAnnouncement
	if err := json.Unmarshal([]byte(job.Payload), &payload); err != nil {
		return PaymentSuccessAnnouncement{}, fmt.Errorf("decode payment announcement payload: %w", err)
	}
	payload.OrderID = strings.TrimSpace(payload.OrderID)
	payload.Provider = strings.ToLower(strings.TrimSpace(payload.Provider))
	payload.ProviderName = strings.TrimSpace(payload.ProviderName)
	payload.Channel = strings.ToLower(strings.TrimSpace(payload.Channel))
	payload.PayableAmount = strings.TrimSpace(payload.PayableAmount)
	payload.PayableCurrency = strings.ToUpper(strings.TrimSpace(payload.PayableCurrency))
	payload.Username = strings.TrimSpace(payload.Username)
	if payload.OrderID == "" || payload.Provider == "" || payload.TXBMinor <= 0 ||
		payload.PayableAmount == "" || payload.PayableCurrency == "" || payload.Username == "" {
		return PaymentSuccessAnnouncement{}, errors.New("payment announcement payload is incomplete")
	}
	return payload, nil
}
