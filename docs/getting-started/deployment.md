# Docker 部署

## 1. 准备环境

需要 Docker Engine 与 Docker Compose v2。生产环境建议配置一个 HTTPS 域名，后端容器在单端口（默认 8080）上同时服务 Webhook、Mini App 页面、API 与静态资源。

只需配置**一个域名**（如 `shop.yourdomain.com`），将 HTTPS 流量转发到 `127.0.0.1:8080` 即可。

## 2. 配置最小 `.env`

复制示例文件：

```bash
cp .env.example .env
```

至少填写：

| 变量 | 说明 |
|------|------|
| `PUBLIC_URL` | 公开 HTTPS 地址，如 `https://shop.yourdomain.com` |
| `BOT_TOKEN` | Telegram Bot Token |
| `ADMIN_IDS` | 管理员 Telegram 用户 ID |
| `WEBAPP_SESSION_SECRET` | 会话密钥（`openssl rand -hex 32`） |
| `POSTGRES_PASSWORD` | 数据库密码 |

> 高级部署如需分离 webhook 与 Mini App 域名，可单独设置 `WEBHOOK_BASE_URL` 与 `SUBSCRIPTION_MINI_APP_URL`（均默认使用 `PUBLIC_URL`）。

> 数据库连接：支持 `SHOP_DATABASE_URL`（User Panel 专属，优先）、`DATABASE_URL`，或从 `POSTGRES_USER` / `POSTGRES_PASSWORD` / `POSTGRES_HOST` / `POSTGRES_PORT` / `POSTGRES_DB` 组件变量自动构建。Redis 支持 `REDIS_URL` 完整连接串，也兼容 `REDIS_HOST` + `REDIS_PORT`。

> 与 Remnawave 面板共用部署时，`db-init` 服务会自动创建两个数据库，无需手动执行 `CREATE DATABASE`。

支付、套餐、Remnawave、外观、语言、汇率建议首次登录后台后配置。`.env` 中的 Remnawave 与支付字段只作为首次启动或后台未覆盖时的兜底值。

## 3. 启动

```bash
docker compose up -d --build
docker compose ps
```

迁移容器会先执行数据库迁移。容器和卷命名统一为 `remna-user-panel-*`。

## 4. 首次后台配置

打开 `https://shop.yourdomain.com` 并使用 Telegram 管理员账号进入后台，然后依次配置：

1. Settings：Webhook Base URL、默认货币、汇率源。
2. Tariffs：套餐目录，默认 USD。
3. Payments：EZPay 或 BEPUSDT。
4. Appearance：Logo、Favicon、主题。
5. Remnawave：面板 API、默认 traffic strategy、squads、HWID 设备限制。

## 5. 设置 Telegram Webhook

```bash
curl -X POST "https://api.telegram.org/bot<BOT_TOKEN>/setWebhook?url=https://shop.yourdomain.com/webhook/telegram"
```

## 6. 其他部署方式

- [与 Remnawave 面板共用部署](deployment-remnawave.md) — 将 User Panel 与 Remnawave 面板部署在同一 docker-compose 中，共用 PostgreSQL 和 Redis
- [Cloudflare Tunnel 部署](deployment-cloudflared.md) — 使用 Cloudflare Tunnel 提供 HTTPS，无需暴露主机端口
