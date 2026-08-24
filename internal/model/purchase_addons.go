package model

import "time"

// PurchaseAddonQuote is the non-mutating, server-priced preview for adding
// paid squads to the member's active purchase.
type PurchaseAddonQuote struct {
	PurchaseID      string    `json:"purchaseId"`
	PriceTXBMinor   int64     `json:"-"`
	Price           Money     `json:"price"`
	ExpiresAt       time.Time `json:"expiresAt"`
	AddonSquadUUIDs []string  `json:"addonSquadUuids"`
}
