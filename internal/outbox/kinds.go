package outbox

// ContinuityKind extends the current upstream entitlement before a queued term
// reaches its local activation boundary.
const ContinuityKind = "remna_prepare_continuity"

// PaymentSuccessAnnouncementKind sends one durable channel announcement for
// the authoritative first transition of a provider payment to paid.
const PaymentSuccessAnnouncementKind = "telegram_payment_success_announcement"

const AffiliateSuccessKind = "telegram_affiliate_success"
const AffiliateTierUpgradeKind = "telegram_affiliate_tier_upgrade"

// UserNotificationKind delivers one immutable private-chat user event.
const UserNotificationKind = "telegram_user_notification"

const (
	UserEventExpiration                 = "expiration"
	UserEventExpiryReminder             = "expiry_reminder"
	UserEventQueuedActivation           = "queued_activation"
	UserEventAutoRenewal                = "auto_renewal"
	UserEventTrafficThreshold           = "traffic_threshold"
	UserEventAutomaticReset             = "automatic_traffic_reset"
	UserEventAutomaticResetInsufficient = "automatic_traffic_reset_insufficient"
	UserEventAutomaticResetFailed       = "automatic_traffic_reset_failed"
	UserEventGroupReward                = "group_reward"
	UserEventAdminExtension             = "admin_extension"
	UserEventAdminUpdate                = "admin_update"
	UserEventNodeCompensation           = "node_compensation"
)
