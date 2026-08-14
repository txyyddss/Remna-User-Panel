package model

import "time"

// AutoRenewal is the current automatic-renewal state and authoritative next-cycle quote.
type AutoRenewal struct {
	PurchaseID       string    `json:"purchaseId"`
	Enabled          bool      `json:"enabled"`
	CanEnable        bool      `json:"canEnable"`
	IneligibleReason *string   `json:"ineligibleReason"`
	GrossPrice       Money     `json:"grossPrice"`
	Discount         Money     `json:"discount"`
	NetPrice         Money     `json:"netPrice"`
	ScheduledAt      time.Time `json:"scheduledAt"`
	NextCycleEndsAt  time.Time `json:"nextCycleEndsAt"`
}

// AutoRenewalFailure is a localized-client-code notice recorded after a due cycle cannot renew.
type AutoRenewalFailure struct {
	PurchaseID string    `json:"purchaseId"`
	Reason     string    `json:"reason"`
	FailedAt   time.Time `json:"failedAt"`
}
