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
- `AdminQuestionnairesPanel.vue` manages questionnaire lifecycles.
- `AdminReasonDialog.vue` collects audited reasons and keeps action failures visible in the dialog.
- `AdminSectionState.vue` renders loading and error states.
- `AdminSettingsPanel.vue` edits runtime settings and specialized calendar/reward controls.
- `AdminShell.vue` provides admin section navigation.
- `AdminShell.test.ts` covers shell navigation and onboarding actions.
- `AdminSquadEditor.vue` edits squad overrides and node access.
- `AdminStatisticsPanel.vue` presents resource statistics.
- `AdminUsersPanel.vue` searches users and adjusts balances.
