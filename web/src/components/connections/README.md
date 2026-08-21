# Connection components

- `ConnectionsPage.vue` composes scans, active blocks, focused dialogs, and cross-refresh after terminal mutations.
- `ConnectionScanStatus.vue` presents the stable starting and polling surface with localized progress, a reduced-motion-safe radar, and no scan mutations.
- `ConnectionNodeList.vue` presents provider nodes, recent IP addresses, and one explicit block command per signed handle.
- `ConnectionBlockList.vue` presents loading, error, empty, active, expiry, and unblock states with light retry/refresh feedback.
- `ConnectionBlockDialog.vue` owns block confirmation plus shared-IP and three-day disclosures; `ConnectionUnblockDialog.vue` owns early-removal confirmation. Close and status refresh controls use light feedback, while destructive block confirmation remains heavy.
- `ConnectionBlockDialog.test.ts` and `ConnectionUnblockDialog.test.ts` cover native Back ownership, warnings, expiry, and confirmation.
- `types.ts` defines the UI-only selected connection target passed from the list to the confirmation surface.
