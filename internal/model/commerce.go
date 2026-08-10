package model

import "time"

// Money is a currency amount encoded without JSON floating point.
type Money struct {
	Currency string `json:"currency"`
	Minor    string `json:"minor"`
	Display  string `json:"display"`
}

// SquadProduct enriches an upstream Remnawave internal squad with customer-facing catalog data.
type SquadProduct struct {
	ID              string    `json:"id"`
	RemnaSquadUUID  string    `json:"remnaSquadUuid"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	PriceTXBMinor   int64     `json:"-"`
	Price           Money     `json:"price"`
	Visible         bool      `json:"visible"`
	UpstreamPresent bool      `json:"upstreamPresent"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// Combo is a time-limited traffic entitlement and its included squads.
type Combo struct {
	ID                      string         `json:"id"`
	Name                    string         `json:"name"`
	Description             string         `json:"description"`
	PriceTXBMinor           int64          `json:"-"`
	Price                   Money          `json:"price"`
	ValidityDays            int            `json:"validityDays"`
	TrafficLimitBytes       int64          `json:"-"`
	TrafficLimit            string         `json:"trafficLimitBytes"`
	ResetStrategy           string         `json:"resetStrategy"`
	Active                  bool           `json:"active"`
	IncludedSquads          []SquadProduct `json:"includedSquads"`
	RolloverMinRemainingBPS int            `json:"rolloverMinRemainingBps"`
	RolloverMaxTXBMinor     int64          `json:"rolloverMaxTxbMinor,string"`
	RolloverMax             Money          `json:"rolloverMax"`
	CreatedAt               time.Time      `json:"createdAt"`
	UpdatedAt               time.Time      `json:"updatedAt"`
}

// Purchase represents an active, queued, or historical combo term.
type Purchase struct {
	ID                      string    `json:"id"`
	UserID                  string    `json:"-"`
	ComboID                 string    `json:"comboId"`
	ComboName               string    `json:"comboName"`
	PriceTXBMinor           int64     `json:"-"`
	Price                   Money     `json:"price"`
	GrossPriceTXBMinor      int64     `json:"-"`
	GrossPrice              Money     `json:"grossPrice"`
	CouponDiscountTXBMinor  int64     `json:"-"`
	CouponDiscount          Money     `json:"couponDiscount"`
	CouponGrantID           *string   `json:"couponGrantId"`
	ValidFrom               time.Time `json:"validFrom"`
	ValidUntil              time.Time `json:"validUntil"`
	Status                  string    `json:"status"`
	TrafficLimitBytes       int64     `json:"-"`
	TrafficLimit            string    `json:"trafficLimitBytes"`
	ResetStrategy           string    `json:"resetStrategy"`
	SquadUUIDs              []string  `json:"squadUuids"`
	RolloverMinRemainingBPS int       `json:"rolloverMinRemainingBps"`
	RolloverMaxTXBMinor     int64     `json:"rolloverMaxTxbMinor,string"`
	RolloverMax             Money     `json:"rolloverMax"`
	CreatedAt               time.Time `json:"createdAt"`
	UpdatedAt               time.Time `json:"updatedAt"`
}

// PurchaseRollover is the durable expiry settlement for one paid term.
type PurchaseRollover struct {
	PurchaseID          string     `json:"purchaseId"`
	Status              string     `json:"status"`
	TrafficLimitBytes   int64      `json:"trafficLimitBytes"`
	UsedTrafficBytes    *int64     `json:"usedTrafficBytes"`
	RemainingBytes      *int64     `json:"remainingTrafficBytes"`
	MinimumRemainingBPS int        `json:"minimumRemainingBps"`
	MaximumTXBMinor     int64      `json:"-"`
	NetPaidTXBMinor     int64      `json:"-"`
	CreditedTXBMinor    int64      `json:"-"`
	ExceptionCode       string     `json:"exceptionCode"`
	Attempts            int        `json:"attempts"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           time.Time  `json:"updatedAt"`
	CompletedAt         *time.Time `json:"completedAt"`
}

// Catalog is the complete customer-visible catalog snapshot.
type Catalog struct {
	Combos []Combo        `json:"combos"`
	Addons []SquadProduct `json:"addons"`
}

// PurchaseQuote is a server-priced, non-mutating checkout preview.
type PurchaseQuote struct {
	ComboID            string    `json:"comboId"`
	ComboName          string    `json:"comboName"`
	GrossPriceTXBMinor int64     `json:"-"`
	DiscountTXBMinor   int64     `json:"-"`
	NetPriceTXBMinor   int64     `json:"-"`
	GrossPrice         Money     `json:"grossPrice"`
	Discount           Money     `json:"discount"`
	NetPrice           Money     `json:"netPrice"`
	EffectiveAt        time.Time `json:"effectiveAt"`
	ExpiresAt          time.Time `json:"expiresAt"`
	Queued             bool      `json:"queued"`
	AddonSquadUUIDs    []string  `json:"addonSquadUuids"`
}
