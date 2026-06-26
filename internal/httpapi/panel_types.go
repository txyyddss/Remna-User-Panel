package httpapi

import "regexp"

const bytesPerGB = 1024 * 1024 * 1024

type paidPlan struct {
	TariffKey         string   `json:"tariff_key"`
	TariffName        string   `json:"tariff_name"`
	Title             string   `json:"title"`
	Description       string   `json:"description"`
	BillingModel      string   `json:"billing_model"`
	SaleMode          string   `json:"sale_mode"`
	Months            int      `json:"months"`
	TrafficGB         float64  `json:"traffic_gb"`
	MonthlyGB         float64  `json:"monthly_gb"`
	SquadUUIDs        []string `json:"squad_uuids"`
	ExternalSquadUUID string   `json:"external_squad_uuid"`
	Provider          string   `json:"provider"`
}

type accessGrantOptions struct {
	Source          string
	TrafficLimitGB  float64
	TrafficStrategy string
	SquadUUIDs      []string
	SetTrafficLimit bool
}

const (
	userTrafficOverridesSettingKey   = "ADMIN_USER_TRAFFIC_OVERRIDES"
	panelSyncStatusSettingKey        = "ADMIN_PANEL_SYNC_STATUS"
	userNotifyExpiryEnabledKey       = "USER_NOTIFY_EXPIRY_ENABLED"
	userNotifyExpiryDaysBeforeKey    = "USER_NOTIFY_EXPIRY_DAYS_BEFORE"
	userNotifyTrafficEnabledKey      = "USER_NOTIFY_TRAFFIC_ENABLED"
	userNotifyTrafficThresholdPctKey = "USER_NOTIFY_TRAFFIC_THRESHOLD_PCT"
	paymentProvisionLockNamespace    = int64(0x5257000000000000)
)

type userTrafficOverride struct {
	PremiumUnlimited  bool   `json:"premium_unlimited_override"`
	PremiumBonusBytes int64  `json:"premium_bonus_bytes"`
	RegularUnlimited  bool   `json:"regular_unlimited_override"`
	RegularBonusBytes int64  `json:"regular_bonus_bytes"`
	UpdatedAt         string `json:"updated_at,omitempty"`
}

// ProvisionResult summarizes one paid-order provisioning pass.
type ProvisionResult struct {
	Scanned     int `json:"scanned"`
	Provisioned int `json:"provisioned"`
	Failed      int `json:"failed"`
}

// PanelSyncResult is persisted for the admin stats page and returned by manual sync.
type PanelSyncResult struct {
	Status              string   `json:"status"`
	LastSyncTime        string   `json:"last_sync_time"`
	UsersProcessed      int      `json:"users_processed"`
	SubscriptionsSynced int      `json:"subscriptions_synced"`
	PaymentsScanned     int      `json:"payments_scanned"`
	PaymentsProvisioned int      `json:"payments_provisioned"`
	PaymentsFailed      int      `json:"payments_failed"`
	Errors              []string `json:"errors,omitempty"`
}

// UserNotificationPrefs holds per-user notification preferences.
type UserNotificationPrefs struct {
	ExpiryEnabled       bool `json:"expiry_enabled"`
	ExpiryDaysBefore    int  `json:"expiry_days_before"`
	TrafficEnabled      bool `json:"traffic_enabled"`
	TrafficThresholdPct int  `json:"traffic_threshold_pct"`
}

const (
	subscriptionNotificationsSentKey = "SUBSCRIPTION_NOTIFICATIONS_SENT"
	subscriptionNotificationsLockKey = "subscription-notification-worker"
)

var panelUsernameInvalid = regexp.MustCompile(`[^A-Za-z0-9_-]+`)

type notifyStage struct {
	key       string
	daysLeft  int
	hoursLeft int
	isExpired bool
	isPostExp bool
}

const (
	defaultExpiryDaysBefore    = 3
	defaultTrafficThresholdPct = 85
)
