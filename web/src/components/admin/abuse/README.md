# Abuse administration UI

- `AbusePolicyCard.vue` edits the revisioned global QPS and continuous-streak policy, validates operational bounds, and emits the complete policy for saving.
- `AbusePunishmentLadder.vue` owns one independently saved escalation row per detector action and spans the desktop admin grid so its controls do not collapse.
- `AbuseRulesCard.vue` manages the create/edit lifecycle for RE2 domain reasons, selects newly created rules after reload, and keeps a compact remote-ID whitelist.
- `AbuseNodesCard.vue` exposes node-key controls and the trailing QPS summary.
- `AbuseRecordsCard.vue` presents responsive, paginated detector records.
- `AbusePolicyCard.test.ts` verifies localized streak loading, bounds, wheel-safe numeric input, and complete-policy submission.
