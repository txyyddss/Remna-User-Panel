# Activity components

Dynamic game and lucky-draw collections use local Motion layout/presence
boundaries. Routine refreshes do not stagger the whole activity surface;
selection and pending feedback remain short and haptic-aware.

- `ActivityPage.vue` composes community activity states in one phone column and a balanced two-column desktop grid.
- `DailyCheckInCard.vue`, `GroupMessageRewardPanel.vue`, `BetGamesPanel.vue`, and `LuckyDrawPanel.vue` implement member activities with selection, confirmation, and authoritative outcome haptics; `GroupMessageRewardPanel.vue` presents message count, reward amount, progress, and claimed/available state in one card.
- `ActivityResultDialog.vue` hosts check-in and bet receipts; `ActivityResultContent.vue` renders shared authoritative reward and balance details, including the selected lucky-draw prize.
- `draw/README.md` specifies the seven result-only presentation modes, focused draw modal, local artwork, and retry/reduced-motion behavior.
- `BetSuccessFireworks.vue` renders a bounded, non-interactive success burst and honors reduced motion.
- `feedback.ts` centralizes result classification and Telegram notification mapping.
- `ActivityResultDialog.test.ts`, `BetSuccessFireworks.test.ts`, and `feedback.test.ts` cover result feedback boundaries and daily check-in reward rendering.
- `gameIcons.ts` maps server-owned icon keys to external Iconify names.
- `ActivityResultDialog.vue`, `DailyCheckInCard.vue`, `GroupMessageRewardPanel.vue`, and `BetSuccessFireworks.vue` own completion-only Motion feedback; particle trajectories remain CSS and reduced motion uses a single success marker.
