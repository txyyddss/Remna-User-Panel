# Admin components

Localized route-level panels and shared admin-only Nuxt UI.

Admin section labels are locale-owned, use Nuxt UI controls, and avoid visual separator literals in component code.
Admin add actions use medium feedback, audited confirmation uses rigid feedback, and destructive confirmation alone uses heavy feedback.

- `AdminActivityPanel.vue` manages games and lucky draws; calendar and reward settings are in `AdminSettingsPanel.vue`.
- `AdminAuditPanel.vue` filters the audit trail.
- `AdminBackupsPanel.vue` manages backup and restore jobs; provider diagnostics stay behind localized status summaries.
- `backups/` contains focused staged-restore and streamed-upload controls.
- `AdminBillingAmountLimits.vue` loads the global inclusive Add TXB range and atomically saves exact minor-unit bounds.
- `AdminCatalogEditor.vue` edits combo terms.
- `AdminCatalogPanel.vue` manages combos and squads.
- `AdminCouponsPanel.vue` manages coupon definitions.
- `AdminDatabasePanel.vue` browses protected database tables.
- `AdminEmbyPanel.vue` monitors Emby provisioning.
- `AdminEntitlementsPanel.vue` manages entitlements.
- `AdminOnboardingPanel.vue` loads, saves, and publishes visual onboarding drafts.
- `onboarding/` contains visual bilingual welcome and agreement card editors.
- `AdminPaymentProfiles.vue` lists and edits masked provider accounts, keeps BEPUSDT channels discovery-owned, deletes profiles through a protected confirmation, and contributes remaining drafts to the parent settings save.
- `AdminQuestionnairesPanel.vue` manages questionnaire lifecycles.
- `AdminReasonDialog.vue` collects audited reasons and keeps action failures visible in the dialog.
- `AdminSectionState.vue` renders loading and error states.
- `AdminSettingsPanel.vue` composes runtime settings, global payment bounds, specialized calendar/reward controls, and one coordinated Save action; payment interventions stay scoped to individual user histories.
- `AdminShell.vue` provides admin section navigation and sends optional signup through Vue Router without native anchor navigation.
- `AdminShell.test.ts` covers shell navigation and onboarding actions.
- `AdminSquadEditor.vue` edits sparse squad overrides; node access remains a live Remnawave projection.
- `admin/squad-profile/` contains the typed Broadband, China Optimized, and International Network profile editor modules.
- `AdminStatisticsPanel.vue` presents resource statistics with narrow-screen overflow protection.
- `AdminUsersPanel.vue` searches users and adjusts balances.
- `users/` contains the aggregate profile, entitlement workflow, operation-resolution, and bulk-extension modules.

The editor also provides dropdowns for coupon-eligible combos and squads, a stock
limit field for squad products, a table-name search for database recovery work,
and mobile-safe backup controls. Balance adjustments use the shared audited
admin endpoint for both credits and deductions.
