package billing

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestPaymentAnnouncementIsBoundedAfterEscaping(t *testing.T) {
	payload := paymentAnnouncementFixture("ezpay", "alipay", "12.34", "CNY")
	payload.Username = strings.Repeat("_", 5000)
	message := formatPaymentSuccessAnnouncement(payload)
	if utf8.RuneCountInString(message) > 4096 || !strings.HasSuffix(message, `\.\.\.`) {
		t.Fatalf("bounded announcement has %d runes and suffix %q", utf8.RuneCountInString(message), message[len(message)-6:])
	}
}
