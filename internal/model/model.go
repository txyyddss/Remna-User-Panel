// Package model defines shared domain values without coupling them to transport or storage.
package model

import (
	"fmt"
	"time"
)

// TelegramProfile is the trusted subset of a validated Mini App user.
type TelegramProfile struct {
	ID        int64  `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Username  string `json:"telegramUsername"`
}

// User is a TX Carpool account.
type User struct {
	ID                   string     `json:"id"`
	TelegramID           int64      `json:"telegramId"`
	TelegramFirstName    string     `json:"firstName"`
	TelegramLastName     string     `json:"lastName"`
	TelegramUsername     string     `json:"telegramUsername"`
	Username             *string    `json:"username"`
	Role                 string     `json:"role"`
	OnboardingState      string     `json:"onboardingState"`
	GroupJoined          bool       `json:"groupJoined"`
	ChannelJoined        bool       `json:"channelJoined"`
	PolicyAcceptedAt     *time.Time `json:"policyAcceptedAt"`
	AgreementRevision    int        `json:"agreementRevision"`
	RemnaUserID          *string    `json:"-"`
	RemnaSubscriptionURL *string    `json:"-"`
	RecoveryReason       string     `json:"recoveryReason"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`
}

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

// LedgerEntry is an immutable TXB balance mutation.
type LedgerEntry struct {
	ID              string    `json:"id"`
	DeltaTXBMinor   int64     `json:"-"`
	Delta           Money     `json:"delta"`
	BalanceAfterRaw int64     `json:"-"`
	BalanceAfter    Money     `json:"balanceAfter"`
	Kind            string    `json:"kind"`
	ReferenceID     string    `json:"referenceId"`
	Note            string    `json:"note"`
	CreatedAt       time.Time `json:"createdAt"`
}

// PaymentOrder is a provider checkout attempt.
type PaymentOrder struct {
	ID                   string     `json:"id"`
	UserID               string     `json:"-"`
	Provider             string     `json:"provider"`
	MethodID             string     `json:"methodId"`
	ProviderRail         string     `json:"providerRail"`
	Status               string     `json:"status"`
	TXBMinor             int64      `json:"-"`
	TXB                  Money      `json:"txb"`
	PayableAmount        string     `json:"payableAmount"`
	PayableCurrency      string     `json:"payableCurrency"`
	RateSnapshot         string     `json:"rateSnapshot"`
	RateDirection        string     `json:"rateDirection"`
	ProviderTradeID      *string    `json:"-"`
	ProviderChargeID     *string    `json:"-"`
	PaymentURL           *string    `json:"paymentUrl"`
	QRPayload            *string    `json:"qrPayload"`
	ReceivingAddress     *string    `json:"receivingAddress"`
	ActualCryptoAmount   *string    `json:"actualCryptoAmount"`
	ActualCryptoCurrency *string    `json:"actualCryptoCurrency"`
	ProviderPayload      string     `json:"-"`
	ExpiresAt            time.Time  `json:"expiresAt"`
	PaidAt               *time.Time `json:"paidAt"`
	RefundedAt           *time.Time `json:"refundedAt"`
	CancelledAt          *time.Time `json:"cancelledAt"`
	CancelReason         string     `json:"cancelReason"`
	ProviderCancelStatus string     `json:"providerCancelStatus"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`
}

// PaymentMethod is one selectable provider rail exposed to a member.
type PaymentMethod struct {
	ID        string `json:"id"`
	Provider  string `json:"provider"`
	Rail      string `json:"rail"`
	Name      string `json:"name"`
	Currency  string `json:"currency"`
	Available bool   `json:"available"`
	Note      string `json:"note"`
}

// Refund is an immutable record of an administrator-authorized payment reversal.
type Refund struct {
	ID             string    `json:"id"`
	PaymentOrderID string    `json:"paymentOrderId"`
	ActorUserID    *string   `json:"actorUserId"`
	TXB            Money     `json:"txb"`
	Reason         string    `json:"reason"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
}

// Statistics is the user-facing subset of Remnawave traffic analytics.
type Statistics struct {
	UsedTrafficBytes     string     `json:"usedTrafficBytes"`
	LifetimeTrafficBytes string     `json:"lifetimeTrafficBytes"`
	TrafficLimitBytes    string     `json:"trafficLimitBytes"`
	OnlineAt             *time.Time `json:"onlineAt"`
	Categories           []string   `json:"categories"`
	SparklineData        []string   `json:"sparklineData"`
	TopNodes             []TopNode  `json:"topNodes"`
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

// StatisticPoint is one timezone-aware aggregate bucket used by accessible
// admin charts and their equivalent data tables.
type StatisticPoint struct {
	PeriodStart string `json:"periodStart"`
	Count       int64  `json:"count"`
	UniqueUsers int64  `json:"uniqueUsers"`
	InputMinor  int64  `json:"inputTxbMinor,string"`
	OutputMinor int64  `json:"outputTxbMinor,string"`
	NetMinor    int64  `json:"netTxbMinor,string"`
}

// StatisticSlice is a labeled count used for win/loss, prize, and squad
// distributions.
type StatisticSlice struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Count int64  `json:"count"`
}

// AdminStatistics is the common detailed-statistics transport for catalog and
// activity resources.
type AdminStatistics struct {
	ResourceID    string           `json:"resourceId"`
	TimeZone      string           `json:"timeZone"`
	From          string           `json:"from"`
	To            string           `json:"to"`
	Bucket        string           `json:"bucket"`
	Count         int64            `json:"count"`
	UniqueUsers   int64            `json:"uniqueUsers"`
	InputMinor    int64            `json:"inputTxbMinor,string"`
	OutputMinor   int64            `json:"outputTxbMinor,string"`
	NetMinor      int64            `json:"netTxbMinor,string"`
	DiscountMinor int64            `json:"discountTxbMinor,string"`
	AddonMinor    int64            `json:"addonTxbMinor,string"`
	Wins          int64            `json:"wins"`
	Losses        int64            `json:"losses"`
	Series        []StatisticPoint `json:"series"`
	Distribution  []StatisticSlice `json:"distribution"`
}

// RemnaNode is the administrator-visible subset of a Remnawave node contract.
type RemnaNode struct {
	UUID                  string   `json:"uuid"`
	Name                  string   `json:"name"`
	CountryCode           string   `json:"countryCode"`
	ConsumptionMultiplier float64  `json:"consumptionMultiplier"`
	ActiveInboundUUIDs    []string `json:"activeInboundUuids"`
	Accessible            bool     `json:"accessible"`
}

// Dashboard combines local financial state with a briefly cached Remnawave snapshot.
type Dashboard struct {
	User              User        `json:"user"`
	Balance           Money       `json:"balance"`
	ActivePurchase    *Purchase   `json:"activePurchase"`
	QueuedPurchase    *Purchase   `json:"queuedPurchase"`
	Statistics        *Statistics `json:"statistics"`
	SubscriptionURL   *string     `json:"subscriptionUrl"`
	StatisticsStale   bool        `json:"statisticsStale"`
	StatisticsWarning string      `json:"statisticsWarning"`
	FetchedAt         time.Time   `json:"fetchedAt"`
}

// TopNode is an upstream node usage summary.
type TopNode struct {
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	CountryCode string `json:"countryCode"`
	TotalBytes  string `json:"totalBytes"`
}

// BackupRun records a database backup attempt.
type BackupRun struct {
	ID          string     `json:"id"`
	Path        string     `json:"path"`
	SizeBytes   int64      `json:"sizeBytes"`
	Status      string     `json:"status"`
	Error       string     `json:"error"`
	CreatedAt   time.Time  `json:"createdAt"`
	CompletedAt *time.Time `json:"completedAt"`
}

// AuditEvent records a privileged mutation.
type AuditEvent struct {
	ID          string    `json:"id"`
	ActorUserID *string   `json:"actorUserId"`
	Action      string    `json:"action"`
	TargetType  string    `json:"targetType"`
	TargetID    string    `json:"targetId"`
	Detail      string    `json:"detail"`
	CreatedAt   time.Time `json:"createdAt"`
}

// OutboxJob is a durable external synchronization operation.
type OutboxJob struct {
	ID          string    `json:"id"`
	Kind        string    `json:"kind"`
	Payload     string    `json:"payload"`
	Status      string    `json:"status"`
	Attempts    int       `json:"attempts"`
	AvailableAt time.Time `json:"availableAt"`
	LastError   string    `json:"lastError"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Setting is a safe administrative view. Encrypted values are never returned.
type Setting struct {
	Key        string    `json:"key"`
	Value      string    `json:"value"`
	Encrypted  bool      `json:"encrypted"`
	Configured bool      `json:"configured"`
	Category   string    `json:"category"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// TXBMoney formats integer hundredths of TXB.
func TXBMoney(minor int64) Money {
	sign := ""
	abs := minor
	if minor < 0 {
		sign = "-"
		abs = -minor
	}
	return Money{Currency: "TXB", Minor: fmt.Sprintf("%d", minor), Display: fmt.Sprintf("%s%d.%02d TXB", sign, abs/100, abs%100)}
}
