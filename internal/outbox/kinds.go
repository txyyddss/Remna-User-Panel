package outbox

// ContinuityKind extends the current upstream entitlement before a queued term
// reaches its local activation boundary.
const ContinuityKind = "remna_prepare_continuity"

// PaymentSuccessAnnouncementKind sends one durable channel announcement for
// the authoritative first transition of a provider payment to paid.
const PaymentSuccessAnnouncementKind = "telegram_payment_success_announcement"

const AffiliateSuccessKind = "telegram_affiliate_success"
const AffiliateTierUpgradeKind = "telegram_affiliate_tier_upgrade"
