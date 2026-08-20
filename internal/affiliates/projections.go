package affiliates

import "time"

type TierProgress struct {
	Current    Tier  `json:"current"`
	Next       *Tier `json:"next,omitempty"`
	Successful int   `json:"successful"`
	Remaining  int   `json:"remaining"`
	TopTier    bool  `json:"topTier"`
}

type Overview struct {
	InviteLink      *string      `json:"inviteLink"`
	TotalCommission Money        `json:"totalCommission"`
	RegisteredCount int          `json:"registeredCount"`
	SuccessfulCount int          `json:"successfulCount"`
	ConversionBPS   int          `json:"conversionBps"`
	TierProgress    TierProgress `json:"tierProgress"`
}

type Referral struct {
	FirstName        string     `json:"firstName"`
	LastName         string     `json:"lastName"`
	RegisteredAt     time.Time  `json:"registeredAt"`
	Status           string     `json:"status"`
	PaybackAt        *time.Time `json:"paybackAt,omitempty"`
	CommissionAmount *Money     `json:"commissionAmount,omitempty"`
}

type ReferralPage struct {
	Items      []Referral `json:"items"`
	Page       int        `json:"page"`
	PageSize   int        `json:"pageSize"`
	Total      int        `json:"total"`
	TotalPages int        `json:"totalPages"`
}

type AdminView struct {
	Version int         `json:"version"`
	Bot     BotIdentity `json:"bot"`
	Tiers   []Tier      `json:"tiers"`
}

type ConfigInput struct {
	ExpectedVersion int    `json:"expectedVersion"`
	Tiers           []Tier `json:"tiers"`
}
