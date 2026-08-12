# Admin components

Localized route-level panels and shared admin-only Nuxt UI.

- `AdminActivityPanel.vue` manages games and lucky draws; calendar and reward settings are in `AdminSettingsPanel.vue`.
- `AdminAuditPanel.vue` filters the audit trail.
- `AdminBackupsPanel.vue` manages backup and restore jobs.
- `AdminCatalogEditor.vue` edits combo terms.
- `AdminCatalogPanel.vue` manages combos and squads.
- `AdminCouponsPanel.vue` manages coupon definitions.
- `AdminDatabasePanel.vue` browses protected database tables.
- `AdminEmbyPanel.vue` monitors Emby provisioning.
- `AdminEntitlementsPanel.vue` manages entitlements.
- `AdminOnboardingPanel.vue` loads, saves, and publishes visual onboarding drafts.
- `onboarding/` contains visual bilingual welcome and agreement card editors.
- `AdminPaymentsPanel.vue` manages payment records, refunds, and safe courtesy credits for failed or expired orders.
- `AdminPaymentProfiles.vue` edits one masked provider profile per EZPay/BEPusdt provider, with independently enabled channels and a custom provider name.
- `AdminQuestionnairesPanel.vue` manages questionnaire lifecycles.
- `AdminReasonDialog.vue` collects audited reasons and keeps action failures visible in the dialog.
- `AdminSectionState.vue` renders loading and error states.
- `AdminSettingsPanel.vue` edits runtime settings and specialized calendar/reward controls.
- `AdminShell.vue` provides admin section navigation and sends optional signup through Vue Router without native anchor navigation.
- `AdminShell.test.ts` covers shell navigation and onboarding actions.
- `AdminSquadEditor.vue` edits squad overrides and node access.
- `AdminStatisticsPanel.vue` presents resource statistics with narrow-screen overflow protection.
- `AdminUsersPanel.vue` searches users and adjusts balances.

The editor also provides dropdowns for coupon-eligible combos and squads, a stock
limit field for squad products, a table-name search for database recovery work,
and mobile-safe backup controls. Balance adjustments use the shared audited
admin endpoint for both credits and deductions.
