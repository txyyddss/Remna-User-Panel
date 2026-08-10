package model

import "time"

// TelegramProfile is the trusted subset of a validated Mini App user.
type TelegramProfile struct {
	ID        int64  `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Username  string `json:"telegramUsername"`
}

// User is a TX Carpool account.
type User struct {
	ID                string     `json:"id"`
	TelegramID        int64      `json:"telegramId"`
	TelegramFirstName string     `json:"firstName"`
	TelegramLastName  string     `json:"lastName"`
	TelegramUsername  string     `json:"telegramUsername"`
	Username          *string    `json:"username"`
	Role              string     `json:"role"`
	OnboardingState   string     `json:"onboardingState"`
	GroupJoined       bool       `json:"groupJoined"`
	ChannelJoined     bool       `json:"channelJoined"`
	PolicyAcceptedAt  *time.Time `json:"policyAcceptedAt"`
	AgreementRevision int        `json:"agreementRevision"`
	RemnaUserID       *string    `json:"-"`
	RecoveryReason    string     `json:"recoveryReason"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}
