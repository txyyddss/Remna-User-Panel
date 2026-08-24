# Abuse administration UI

- `AbusePolicyCard.vue` edits the revisioned global QPS policy and validates operational bounds before a save.
- `AbusePunishmentLadder.vue` owns one independently saved escalation row per detector action.
- `AbuseRulesCard.vue` manages RE2 domain reasons and a compact remote-ID whitelist.
- `AbuseNodesCard.vue` exposes node-key controls and the trailing QPS summary.
- `AbuseRecordsCard.vue` presents responsive, paginated detector records.
