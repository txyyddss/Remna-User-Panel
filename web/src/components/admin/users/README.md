# Admin user workflows

- `AdminBulkExtensionDialog.vue` previews inclusive-OR active targets and queues the audited durable extension job.
- `AdminComboReplacementDialog.vue` changes the active combo and optional squads without moving TXB.
- `AdminEntitlementEditor.vue` validates the advanced optimistic-lock overwrite form.
- `AdminOperationResolutionDialog.vue` records an audited resolution without retrying an ambiguous provider mutation.
- `AdminUserEntitlements.vue` presents active and queued entitlements with edit, refund, and replacement actions.
- `AdminUserHistory.vue` presents Emby accounts, payments, refunds, and open provider operations.
- `AdminUserOverview.vue` presents identity, balance, synchronization, and the active combo.
- `AdminUserProfilePage.vue` composes the aggregate user workflow and its focused dialogs.
- `adminUserFormat.ts` owns profile-specific status tones and date conversion helpers.
- `useAdminUserProfile.ts` loads the aggregate and serializes idempotent profile mutations with conflict handling.
