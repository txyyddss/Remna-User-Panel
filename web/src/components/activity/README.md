# Activity components

- `ActivityPage.vue` composes community activity states.
- `DailyCheckInCard.vue`, `GroupMessageRewardPanel.vue`, `BetGamesPanel.vue`, and `LuckyDrawPanel.vue` implement member activities.
- `ActivityResultDialog.vue` presents authoritative results and mounts success feedback only for winning bets.
- `BetSuccessFireworks.vue` renders a bounded, non-interactive success burst and honors reduced motion.
- `feedback.ts` centralizes result classification and Telegram notification mapping.
- `ActivityResultDialog.test.ts`, `BetSuccessFireworks.test.ts`, and `feedback.test.ts` cover result feedback boundaries.
- `gameIcons.ts` maps server-owned icon keys to external Iconify names.
