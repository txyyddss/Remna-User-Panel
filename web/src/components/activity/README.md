# Activity components

- `ActivityPage.vue` composes community activity states in one phone column and a balanced two-column desktop grid.
- `DailyCheckInCard.vue`, `GroupMessageRewardPanel.vue`, `BetGamesPanel.vue`, and `LuckyDrawPanel.vue` implement member activities; `GroupMessageRewardPanel.vue` presents message count, reward amount, progress, and claimed/available state in one card.
- `ActivityResultDialog.vue` presents authoritative results, including the exact TXB amount added by a daily check-in and rewards returned by lucky draws, and mounts success feedback only for winning bets.
- `BetSuccessFireworks.vue` renders a bounded, non-interactive success burst and honors reduced motion.
- `feedback.ts` centralizes result classification and Telegram notification mapping.
- `ActivityResultDialog.test.ts`, `BetSuccessFireworks.test.ts`, and `feedback.test.ts` cover result feedback boundaries and daily check-in reward rendering.
- `gameIcons.ts` maps server-owned icon keys to external Iconify names.
