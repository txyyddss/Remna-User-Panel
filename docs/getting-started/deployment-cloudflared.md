# Cloudflare Tunnel 部署

使用 Cloudflare Tunnel（cloudflared）为 Remna User Panel 提供 HTTPS 反向代理，无需在主机上暴露任何端口。Cloudflared 通过出站连接建立加密隧道，自动处理 TLS 证书。

**只需一个域名**，后端单端口同时服务 webhook、Mini App 页面、API 和静态资源。

## 优势

- **零端口暴露**：主机防火墙无需开放任何入站端口
- **自动 HTTPS**：Cloudflare 自动提供和管理 TLS 证书
- **DDoS 防护**：Cloudflare 的全球网络提供 DDoS 缓解
- **单域名部署**：不再需要分离 webhook 和 Mini App 域名

## 架构

```
 ┌──────────────────────────────────────────────────┐
 │                  Cloudflare                        │
 │        shop.yourdomain.com                        │
 └──────────────────────┬───────────────────────────┘
                        │  Cloudflare Tunnel
                        │  (出站加密连接)
 ┌──────────────────────┴───────────────────────────┐
 │                    Docker Host                     │
 │  ┌────────────────────────────────────────────┐   │
 │  │  cloudflared (出站 → Cloudflare)            │   │
 │  └──────────────────┬─────────────────────────┘   │
 │                     │                              │
 │  ┌──────────────────┴─────────────────────────┐   │
 │  │  backend :8080                              │   │
 │  │  (webhook + Mini App + API + 静态资源)       │   │
 │  └──────────────────┬─────────────────────────┘   │
 │                     │                              │
 │  ┌──────────────────┼─────────────────────────┐   │
 │  │  postgres :5432  │  redis :6379            │   │
 │  └──────────────────┴─────────────────────────┘   │
 └───────────────────────────────────────────────────┘
```

## 1. 前置条件

- 一个由 **Cloudflare 管理 DNS** 的域名
- 有效的 Cloudflare 账号
- Docker Engine 与 Docker Compose v2

## 2. 创建 Cloudflare Tunnel

1. 登录 [Cloudflare Zero Trust](https://one.dash.cloudflare.com/)
2. 进入 **Networks → Tunnels**
3. 点击 **Create a tunnel**
4. 命名（如 `remna-user-panel`），选择环境（Docker），复制 **Tunnel Token**
5. 添加公共主机名（**Public Hostnames**）：

| 子域名 | 服务 |
|--------|------|
| `shop.yourdomain.com` | `http://backend:8080` |

> 后端单端口服务所有内容 — webhook、Mini App、API 和静态资源都通过这一个主机名访问。

## 3. 配置环境变量

```bash
cd deploy/examples/cloudflared
cp .env.example .env
```

编辑 `.env`，填写：

```dotenv
# Cloudflare Tunnel Token（从 Zero Trust 复制）
CF_TUNNEL_TOKEN=eyJhIjoi...

# 单个公开 URL — 必须与 Tunnel 中配置的公共主机名一致
PUBLIC_URL=https://shop.yourdomain.com

# Telegram
BOT_TOKEN=your_bot_token_here
ADMIN_IDS=123456789

# 安全密钥
WEBAPP_SESSION_SECRET=...
WEBHOOK_SECRET_TOKEN=...

# Remnawave 面板
PANEL_API_URL=https://your-panel.yourdomain.com/api
PANEL_API_KEY=...
```

## 4. 启动

```bash
docker compose up -d
docker compose ps
```

所有服务通过内部 Docker 网络通信，无需暴露任何主机端口。

## 5. 设置 Telegram Webhook

```bash
curl -X POST "https://api.telegram.org/bot<BOT_TOKEN>/setWebhook?url=https://shop.yourdomain.com/webhook/telegram"
```

## 6. 首次后台配置

1. 打开 `https://shop.yourdomain.com`，用 Telegram 管理员账号登录
2. 在后台依次配置：Settings → Tariffs → Payments → Appearance → Remnawave

## 使用配置文件替代 Token

如果你更喜欢用配置文件管理 Tunnel：

1. 在 Cloudflare Zero Trust 创建 Tunnel 后，下载 `credentials.json`
2. 将 `credentials.json` 放入 `cloudflared/` 目录
3. 复制 `cloudflared/config.yml.example` 为 `cloudflared/config.yml`，修改 `tunnel` UUID 和 `hostname`
4. 修改 `docker-compose.yml` 中 cloudflared 服务的 command 和 volumes：

```yaml
cloudflared:
  image: cloudflare/cloudflared:latest
  command: tunnel --no-autoupdate run
  volumes:
    - ./cloudflared:/etc/cloudflared:ro
  networks:
    - remna-user-panel
```

## 故障排查

- **Tunnel 无法连接**：检查 `CF_TUNNEL_TOKEN` 是否正确，Tunnel 在 Zero Trust 中是否为 Active 状态
- **404 错误**：确认公共主机名的 Service URL 为 `http://backend:8080`（注意是 Docker 容器名，不是 `localhost`）
- **Webhook 不工作**：确认 `PUBLIC_URL` 与 Tunnel 的公共主机名完全一致
- **客户端 IP 不正确**：确保 `TRUSTED_PROXIES` 包含 Docker 网络子网（默认已包含 `10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16`）
