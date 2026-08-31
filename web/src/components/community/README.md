# Community components

- `CommunityPage.vue` orchestrates canonical membership loading, activation refreshes, localized feedback, and per-space invite actions.
- `CommunityMembershipRows.vue` is the presentational two-row group and channel access surface; confirmed membership takes precedence over eligibility status.
- `CommunityPage.test.ts` covers loading, error fallback, and canonical row projection; `CommunityMembershipRows.test.ts` covers Joined precedence and per-space join intent.
