# REST API Reference

The backend exposes a wide range of APIs structured under `/api/v1/`. The API follows RESTful principles and mainly uses JSON payloads.

## Authentication

Several endpoints require an active session. The backend uses the `CombinedAuth` middleware which supports authenticating via:
- Standard session tokens/cookies for the Web UI.
- `X-Telegram-Init-Data` header for Telegram Mini App contexts.

There is also an `AdminOnly` middleware required for specific endpoints that manage configurations and combo templates.

---

## Endpoint Groups

### Public Endpoints
- **GET `/health`**
  Check the API status and server time.
- **POST `/api/v1/payment/callback/bepusdt`**
  Webhook endpoint for BEPusdt asynchronous payment completions.
- **POST `/api/v1/payment/callback/ezpay`**
  Webhook endpoint for EZPay callbacks.

---

### User Profile
*Requires `CombinedAuth`*
- **GET `/api/v1/user/me`**
  Fetch current user profile and linked third-party UUIDs.

---

### Credit System (TXB)
*Requires `CombinedAuth`*
- **GET `/api/v1/credit/balance`**
  Retrieve the current TXB credit limit and balance.
- **POST `/api/v1/credit/signup`**
  Claim registration bonuses in TXB credits.
- **POST `/api/v1/credit/bet`**
  Engage in the TXB credit betting minigame.
- **GET `/api/v1/credit/history`**
  Paginated listing of credit transaction history.

---

### Combos (Plans)
*Requires `CombinedAuth`*
- **GET `/api/v1/combos`**
  List all active combos available for users to purchase.

---

### Remnawave VPN Subscriptions
*Requires `CombinedAuth`*
- **POST `/api/v1/subscribe`**
  Purchase or renew a VPN combo (Remnawave subscription). Uses available credits, balances, or calculates outstanding gaps.
- **GET `/api/v1/sub-info`**
  Retrieve active VPN subscription status and traffic limits.
- **GET `/api/v1/sub-keys`**
  Retrieve subscription links for compatible client software (e.g. V2Ray, Clash).
- **GET `/api/v1/vpn/bandwidth`**
  View statistics on node bandwidth utilization.
- **GET `/api/v1/vpn/devices`**
  Retrieve the maximum device limits and currently active device information based on HWID constraints.
- **GET `/api/v1/vpn/ips`**
  Check active IPv4/IPv6 sessions currently bounded to this subscription under Remnawave.
- **GET `/api/v1/vpn/history`**
  Obtain history logs for past subscriptions and renewals.

---

### VPN Node & IP Settings
*Requires `CombinedAuth`*
- **POST `/api/v1/ip/change`**
  Force drop all active sessions on Remnawave core to obtain a new IP address, restricted by the cooldown config.
- **GET `/api/v1/ip/status`**
  Check current IP change cooldown status (allows IP refresh if available).
- **GET `/api/v1/squads/external`**
  Fetch accessible external node squads.
- **PUT `/api/v1/squads/external`**
  Modify active group squads for connection.

---

### Payments
*Requires `CombinedAuth`*
- **POST `/api/v1/payment/create`**
  Initialize a deposit/recharge request returning gateway URLs (either BEPusdt or EZPay).

---

### Jellyfin
*Requires `CombinedAuth`*
- **POST `/api/v1/jellyfin/purchase`**
  Register a Jellyfin media account leveraging base Remna balance.
- **POST `/api/v1/jellyfin/quick-connect`**
  Generate a pin login code utilizing Jellyfin Quick Connect protocol via the SDK.
- **PUT `/api/v1/jellyfin/password`**
  Update Jellyfin password via API bypassing frontend requirement.
- **GET `/api/v1/jellyfin/devices`**
  Fetch active Jellyfin streaming devices.
- **PUT `/api/v1/jellyfin/parental-rating`**
  Adjust allowed age ratings and media availability bounds.

---

### Admin Actions
*Requires `AdminOnly` middleware*
- **GET `/api/v1/admin/config`**
  Retrieve full server configuration settings dynamically.
- **PUT `/api/v1/admin/config`**
  Modify config at runtime. Hot-reload will be triggered.
- **POST `/api/v1/admin/combos`**
  Create a new subscription plan template.
- **PUT `/api/v1/admin/combos/{uuid}`**
  Edit an existing plan template.
- **GET `/api/v1/admin/squads/internal`**
  Fetch standard Remnawave node groups/squads to configure a combo template.
