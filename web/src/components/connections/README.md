# Connection components

- `ConnectionsPage.vue` composes scan, progress, empty, failure, retry, and durable drop-receipt states for the member route.
- `ConnectionScanStatus.vue` presents the stable starting and polling surface with localized progress, a reduced-motion-safe radar, and no scan mutations.
- `ConnectionNodeList.vue` presents provider nodes, country flags, recent IP addresses, last-seen timestamps, and one explicit unlink command per signed handle.
- `ConnectionDropDialog.vue` owns selected-IP confirmation, shared-IP collateral disclosure, operation status, and native Telegram Back while visible.
- `ConnectionDropDialog.test.ts` verifies that the visible confirmation takes native Back ownership and closes through its model contract.
- `types.ts` defines the UI-only selected connection target passed from the list to the confirmation surface.
