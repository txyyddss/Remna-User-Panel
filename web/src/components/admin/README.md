# Admin components

Localized route-level panels and shared admin-only Nuxt UI.

Admin section labels are locale-owned, use Nuxt UI controls, and avoid visual separator literals in component code.
Admin controls inherit guarded action feedback, modal dismissal stays soft, audited confirmation uses rigid feedback, and destructive entry or confirmation uses heavy feedback.

- `AdminActivityPanel.vue` manages games and lucky draws; calendar and reward settings are in `AdminSettingsPanel.vue`.
- `AdminAuditPanel.vue` filters the audit trail.
- `AdminBackupsPanel.vue` manages backup and restore jobs; future continuity work shows its localized scheduled state and eligible time while provider diagnostics stay behind status summaries.
- `backups/` contains focused staged-restore and streamed-upload controls.
- `AdminBillingAmountLimits.vue` loads the global inclusive Add TXB range and atomically saves exact minor-unit bounds.
- `AdminCatalogEditor.vue` edits combo terms.
- `AdminCatalogPanel.vue` manages combos and squads.
- `AdminCompensationPanel.vue` composes the revisioned outage policy and recipient-safe event review flow, with an explicitly left-aligned localized page title and subtitle.
- `AdminAbusePanel.vue` lazy-loads the premium-dark detector policy, node-key, rule, statistics, and record surfaces.
- `abuse/` contains focused detector administration cards.
- `AdminCouponsPanel.vue` manages coupon definitions.
- `AdminDatabasePanel.vue` browses protected database tables.
- `AdminEmbyPanel.vue` monitors Emby provisioning.
- `AdminEntitlementsPanel.vue` manages entitlements.
- `AdminOnboardingPanel.vue` loads, saves, and publishes visual onboarding drafts.
- `onboarding/` contains visual bilingual welcome and agreement card editors.
- `AdminPaymentProfiles.vue` lists and edits masked provider accounts, keeps BEPUSDT channels discovery-owned, deletes profiles through a protected confirmation, and contributes remaining drafts to the parent settings save.
- `AdminQuestionnairesPanel.vue` manages questionnaire lifecycles.
- `AdminReasonDialog.vue` collects audited reasons and keeps action failures visible in the dialog.
- `AdminSectionNavigation.vue` remains the grouped section-control building block; `AdminShell.vue` keeps only the mobile selector because the desktop sidebar owns administrator section navigation.
- `AdminSectionState.vue` renders loading and error states.
- `AdminSettingsPanel.vue` composes runtime settings, global payment bounds, specialized calendar/reward controls, and one coordinated Save action; payment interventions stay scoped to individual user histories.
- `AdminShell.vue` provides the mobile admin section selector and sends optional signup through Vue Router without native anchor navigation.
- `AdminShell.test.ts` covers shell navigation and onboarding actions.
- `AdminSquadEditor.vue` edits sparse squad overrides; node access remains a live Remnawave projection.
- `admin/squad-profile/` contains the typed Broadband, China Optimized, and International Network profile editor modules.
- `AdminStatisticsPanel.vue` presents resource statistics with narrow-screen overflow protection.
- `AdminUsersPanel.vue` searches users with responsive state/combo/squad facets and adjusts balances.
- `users/` contains the aggregate profile, wallet, provider-account, entitlement workflow, operation-resolution, and bulk-extension modules.
- `compensation/` contains focused configuration, event-card, review-modal, and display-format modules.

The editor also provides dropdowns for coupon-eligible combos and squads, a stock
limit field for squad products, a table-name search for database recovery work,
and mobile-safe backup controls. Balance adjustments use the shared audited
admin endpoint for both credits and deductions.
