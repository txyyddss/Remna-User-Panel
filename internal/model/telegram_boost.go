package model

// TelegramBoostAppreciation contains only the identity needed by a queued thank-you.
type TelegramBoostAppreciation struct {
	ChatID   int64  `json:"chatId"`
	BoostID  string `json:"boostId"`
	Username string `json:"username,omitempty"`
	Name     string `json:"name,omitempty"`
	Locale   string `json:"locale,omitempty"`
}
