# Internal Go SDK Documentation

The backend service integrates primarily with four disparate external services. Interaction with each API interface is abstracted into strongly-typed logic within `backend/internal/sdk/`.

## 1. Remnawave SDK (`/sdk/remnawave`)
The VPN subscription operations and user synchronizations are backed by the [Remnawave Panel](https://github.com/remnawave/remnawave).
- **Connection Management**: Methods configured like `DropConnections()` issue drop operations directly interacting with core routing processes to sever existing proxies, useful during **IP Refresh**.
- **User Subscription Generation**: Synchronizing keys and quotas on each order payment.
- **Bandwidth Metrics**: Data scraping to calculate remaining user traffic limits iteratively against the overall panel limit.

## 2. Jellyfin SDK (`/sdk/jellyfin`)
Powers media server synchronization. It targets a privately hosted Jellyfin setup.
- **QuickConnect Code**: Automatically creates and fetches pins so users can sign into smart TVs seamlessly without manual typing mapping to `JellyfinQuickConnect` functionality.
- **Account Registration & Parental Control Policy**: `CreateUser` limits their viewing based on the selected package level. Directly injects data to Jellyfin’s policy databases synchronously.
- **Device Management**: Allows killing connections for users overriding device constraints.

## 3. BEPusdt SDK (`/sdk/bepusdt`)
Cryptocurrency automated dual-payment system on Binance Smart Chain tracking natively TRC20/BEP20.
- **Order Tracking API**: Allows issuing pending block scanning hashes which trigger webhooks when coins hit the blockchain.

## 4. EZPay SDK (`/sdk/ezpay`)
Traditional fiat processor gateway implementation bridging Alipay/WeChat to equivalent User-Panel tokens using the generalized "EPC/VMP" standardized APIs.
- Wraps sign strings, digests query params via MD5 logic according to EZPay specs for creating verifiable order URIs.
