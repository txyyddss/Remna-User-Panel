# 后台配置

管理后台负责运行时配置；若某个同名变量明确存在于 `.env`，环境值优先，后台字段会锁定且服务端拒绝覆盖。

## 环境变量参考

### 数据库连接（三选一，优先级从高到低）

| 变量 | 说明 |
|------|------|
| `SHOP_DATABASE_URL` | User Panel 专属数据库连接串，优先于 `DATABASE_URL` |
| `DATABASE_URL` | 通用数据库连接串（注意：与 Remnawave 面板共享 .env 时请使用 `SHOP_DATABASE_URL`） |
| `POSTGRES_USER` / `POSTGRES_PASSWORD` / `POSTGRES_HOST` / `POSTGRES_PORT` / `POSTGRES_DB` | 组件变量，当以上 URL 均未设置时自动构建连接串 |

### Redis 连接

| 变量 | 说明 |
|------|------|
| `REDIS_URL` | Redis 完整连接串，如 `redis://redis:6379/0` |
| `REDIS_HOST` + `REDIS_PORT` | 当 `REDIS_URL` 未设置时，自动构建 |

### Web UI 可管理设置

以下配置可在管理后台 → Settings 页面直接修改，无需编辑 `.env`：

| 设置键 | 说明 | 默认值 |
|--------|------|--------|
| `DEFAULT_CURRENCY_SYMBOL` | 默认目录币种 | USD |
| `DEFAULT_LANGUAGE` | 默认语言 | zh |
| `USER_SQUAD_UUIDS` | 默认内部 Squad | （空） |
| `USER_EXTERNAL_SQUAD_UUID` | 默认外部 Squad | （空） |
| `USER_HWID_DEVICE_LIMIT` | 默认 HWID 设备限制 | （空，使用 Remnawave 默认） |
| `SUBSCRIPTION_NOTIFY_DAYS_BEFORE` | 到期前通知天数 | 3 |
| `SUBSCRIPTION_NOTIFY_HOURS_BEFORE` | 到期前通知小时数 | 0 |
| `WORKER_PANEL_SYNC_INTERVAL_SECONDS` | 面板同步间隔（需重启） | 900 |
| `WORKER_PAYMENT_PROVISION_INTERVAL_SECONDS` | 支付处理间隔（需重启） | 30 |
| `TELEMETRY_ENABLED` | 本机匿名遥测与奖励风控 | true |
| `TELEMETRY_RETENTION_HOURS` | 无命中匿名数据保留小时数（1-720） | 24 |
| `TELEMETRY_FINGERPRINT_REJECT_SCORE` | 欢迎奖励指纹拒绝阈值（1-100） | 70 |
| `TELEGRAM_LOGIN_CLIENT_ID` | BotFather Web Login Client ID | （空） |

> 以上键名也支持通过环境变量强制设置。只要变量存在（包括空值），后台即锁定该字段。

## 支付与回调

在 Settings 中填写 `WEBHOOK_BASE_URL`，格式为后端公网 HTTPS 地址，例如：

```text
https://api.example.com
```

系统会生成：

- EZPay：`https://api.example.com/webhook/ezpay`
- BEPUSDT：`https://api.example.com/webhook/bepusdt`

EZPay 与 BEPUSDT 的 return URL 自动使用 `SUBSCRIPTION_MINI_APP_URL`，不再单独配置。Telegram Stars 通过 `STARS_ENABLED` 开启，并要求套餐存在 `stars_price`。

## 浏览器 Telegram 登录

在 BotFather 的 **Bot Settings → Web Login** 登记站点 Allowed URL，并将 Client ID 写入 `TELEGRAM_LOGIN_CLIENT_ID`。浏览器使用新版 Login Library 和服务端 nonce/JWKS 校验；Telegram Mini App 继续使用 `initData`。

## 本机遥测和欢迎奖励风控

遥测只写入本机 PostgreSQL，不发送到外部。浏览器指纹组件会先在浏览器摘要，再由服务端 HMAC；数据库不保存原始 Canvas、音频、字体、IP 或 User-Agent。清理任务每小时运行，超过 `TELEMETRY_RETENTION_HOURS` 未命中的匿名记录会被删除。

## 汇率

默认 provider 是 `frankfurter`。也可以选择：

- `exchange_rate_api`
- `custom`

汇率会缓存，provider 失败时使用最后一次有效缓存；没有缓存时使用保守 fallback，并在后台状态中标记为 stale。

## 语言与外观

语言文件位于 `locales/zh.json` 与 `locales/en.json`。管理端不再提供在线翻译页面，现有 override 文件仍兼容读取。

Logo、Favicon 和主题都在 Appearance 中维护。
