package providerops

const (
	// KindHostRemarkUpdate synchronizes one scheduled multiplier token update.
	KindHostRemarkUpdate = "host_remark_update"
	// KindAdminEntitlementEdit synchronizes one audited full-field edit.
	KindAdminEntitlementEdit = "admin_entitlement_edit"
	// KindAdminEntitlementRefund synchronizes one credited cancellation.
	KindAdminEntitlementRefund = "admin_entitlement_refund"
	// KindAdminComboReplacement synchronizes one no-charge configuration change.
	KindAdminComboReplacement = "admin_combo_replacement"
	// KindAdminBulkExtension synchronizes deduplicated active-user extensions.
	KindAdminBulkExtension = "admin_bulk_extension"
	// KindNodeCompensation synchronizes one reviewed outage compensation.
	KindNodeCompensation = "node_compensation"
)
