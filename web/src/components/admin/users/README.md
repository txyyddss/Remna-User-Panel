# Admin user workflows

- `AdminBulkExtensionDialog.vue` previews inclusive-OR active targets and queues the audited durable extension job.
- `AdminDurationField.vue` and `duration.ts` provide the minute-normalized amount/unit control shared by administrative extension workflows.
- `duration.test.ts` covers deterministic minute, hour, and day conversion boundaries.
- `AdminComboReplacementDialog.vue` changes the active combo and optional squads without moving TXB.
- `AdminEntitlementEditor.vue` validates the advanced optimistic-lock overwrite form.
- `AdminOperationResolutionDialog.vue` records an audited resolution without retrying an ambiguous provider mutation.
- `AdminUserEntitlements.vue` presents active and queued entitlements with edit, refund, and replacement actions.
- `AdminUserHistory.vue` presents Emby accounts, payments, refunds, and open provider operations, with refund and courtesy-credit actions only for that profile's eligible payment records.
- `AdminUserIPBlocks.vue` presents active member-created blocks and exposes only the administrator unblock action.
- `AdminUserOverview.vue` presents identity, balance, synchronization, and the active combo.
- `AdminUserAccountContext.vue` presents the active coupon wallet, retained abuse records, and affiliate history without duplicating member facts.
- `AdminUserProfilePage.vue` composes the aggregate user workflow and its focused dialogs.
- `AdminUserIPBlocks.test.ts` verifies that administrator profiles expose unblock without a block control.
- `adminUserFormat.ts` owns profile-specific status tones and date conversion helpers.
- `useAdminUserProfile.ts` loads the aggregate and serializes idempotent profile mutations with conflict handling.
