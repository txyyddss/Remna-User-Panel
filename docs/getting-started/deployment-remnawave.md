# 与 Remnawave 面板共用部署

本指南介绍如何将 Remna User Panel 与 Remnawave 面板部署在同一台服务器上，共用 PostgreSQL 和 Redis 实例。只需**一个域名**即可。

## 架构

```
┌─────────────────────────────────────────────────────┐
│                    Docker Host                       │
│  ┌──────────┐  ┌──────────┐  ┌───────────────────┐ │
│  │ Postgres │  │  Redis   │  │  Reverse Proxy    │ │
│  │   :5432  │  │  :6379   │  │  :80/:443         │ │
│  └────┬─────┘  └────┬─────┘  └────────┬──────────┘ │
│       │             │                 │             │
│  ┌────┴─────────────┴─────────────────┴──────────┐  │
│  │              remnawave-net                    │  │
│  │  ┌────────────────┐  ┌────────────────────┐  │  │
│  │  │ remnawave-panel│  │ user-panel-backend │  │  │
│  │  │    :3000       │  │    :8080           │  │  │
│  │  └────────────────┘  └────────────────────┘  │  │
│  └──────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────┘
```

后端单端口（8080）同时服务 webhook、Mini App 页面、API 和静态资源。只需一个域名（如 `shop.yourdomain.com`）即可。

## 1. 准备

- Docker Engine 与 Docker Compose v2
- 一个域名（可由 Cloudflare 管理 DNS，推荐用于 Cloudflare Tunnel）
- Remnawave 面板已安装或准备安装

## 2. 配置环境变量

```bash
cd deploy/examples/remnawave
cp .env.example .env
```

编辑 `.env`，至少填写：

| 变量 | 说明 | 所属 |
|------|------|------|
| `POSTGRES_PASSWORD` | PostgreSQL 超级用户密码 | 共享 |
| `PUBLIC_URL` | 公开 HTTPS 地址，如 `https://shop.yourdomain.com` | User Panel |
| `BOT_TOKEN` | Telegram Bot Token（店铺机器人） | User Panel |
| `ADMIN_IDS` | 管理员 Telegram 用户 ID | User Panel |
| `ADMIN_EMAIL` | 可选：管理员密码登录邮箱 | User Panel |
| `ADMIN_PASSWORD` | 可选：管理员密码登录密码 | User Panel |
| `WEBAPP_SESSION_SECRET` | 会话密钥（`openssl rand -hex 32`） | User Panel |
| `WEBHOOK_SECRET_TOKEN` | Webhook 密钥（`openssl rand -hex 32`） | User Panel |
| `PANEL_API_URL` | Remnawave 面板 API 地址，同一网络内可用 `http://remnawave-panel:3000` | User Panel |
| `PANEL_API_KEY` | Remnawave 面板 API 密钥 | User Panel |
| `JWT_AUTH_SECRET` | JWT 认证密钥（`openssl rand -hex 64`） | Remnawave 面板 |
| `JWT_API_TOKENS_SECRET` | JWT API 令牌密钥（`openssl rand -hex 64`） | Remnawave 面板 |
| `SUB_PUBLIC_DOMAIN` | 订阅公开域名，如 `sub.yourdomain.com` | Remnawave 面板 |

> `.env.example` 中已列出所有可选变量（Telegram 通知、支付、HWID 限制等），按需取消注释即可。

### 变量命名空间

两个项目的变量在 `.env` 中自然共存，互不冲突：

- **共享基础设施**：`POSTGRES_USER`、`POSTGRES_PASSWORD`（PostgreSQL 容器使用）
- **数据库隔离**：`REMAWAVE_DB`、`USER_PANEL_DB` 分别指定两个项目的数据库名，`db-init` 服务自动创建
- **数据库连接**：`docker-compose.yml` 为每个服务单独注入 `DATABASE_URL`，User Panel 还支持 `SHOP_DATABASE_URL` 作为专属变量（优先于 `DATABASE_URL`）
- **Redis**：User Panel 使用 `REDIS_URL`（完整连接串），也兼容 Remnawave 风格的 `REDIS_HOST` + `REDIS_PORT`
- **Web UI 管理**：默认语言、货币、Squad、HWID 限制、通知天数、支付等配置请在后台 Settings 页面修改

## 3. 创建数据库

首次启动时 `db-init` 服务会自动创建 `remnawave` 和 `remna_user_panel` 两个数据库，无需手动操作。

```bash
docker compose up -d
```

如需要手动创建：

```bash
docker compose up -d postgres redis
docker compose exec postgres psql -U postgres -c "CREATE DATABASE remnawave;"
docker compose exec postgres psql -U postgres -c "CREATE DATABASE remna_user_panel;"
docker compose up -d
```

## 4. 启动

```bash
docker compose up -d
docker compose ps
```

## 5. 反向代理

单个域名指向后端即可。推荐方案：

### Cloudflared（推荐，无需暴露端口）

参见 [Cloudflare Tunnel 部署指南](deployment-cloudflared.md)。

在 Cloudflare Zero Trust 中为 Tunnel 添加公共主机名：

| 公共主机名 | 目标服务 |
|-----------|---------|
| `shop.yourdomain.com` | `http://user-panel-backend:8080` |
| `panel.yourdomain.com` | `http://remnawave-panel:3000` |

### 外部 Nginx / Caddy

```nginx
# Nginx 反向代理 — 单域名
upstream user_panel {
    server 127.0.0.1:8080;
}

server {
    listen 443 ssl;
    server_name shop.yourdomain.com;
    location / { proxy_pass http://user_panel; }
}
```

## 6. 首次后台配置

1. 打开 `https://shop.yourdomain.com`，用 Telegram 管理员账号登录
2. 在后台依次配置：Settings → Tariffs → Payments → Appearance → Remnawave

## 7. 设置 Telegram Webhook

```bash
curl -X POST "https://api.telegram.org/bot<BOT_TOKEN>/setWebhook?url=https://shop.yourdomain.com/webhook/telegram"
```

## 故障排查

- **数据库连接失败**：检查 `POSTGRES_HOST=postgres` 和网络 `remnawave-net`
- **面板 API 不通**：确认 `PANEL_API_URL` 使用 Docker 容器名（如 `http://remnawave-panel:3000`）
- **Webhook 不工作**：确认 `PUBLIC_URL` 可从公网访问，且 Telegram 服务器能连接
