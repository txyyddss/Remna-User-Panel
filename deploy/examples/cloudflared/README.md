# Cloudflare Tunnel 部署

详细文档见：[docs/getting-started/deployment-cloudflared.md](../../../docs/getting-started/deployment-cloudflared.md)

## 文件说明

- `docker-compose.yml`：使用 Cloudflare Tunnel 的完整部署（无需暴露主机端口）
- `.env.example`：最小环境变量模板
- `cloudflared/config.yml.example`：使用配置文件方式替代 Token 的参考

## 快速开始

```bash
cp .env.example .env
# 编辑 .env：
#   1. 填写 CF_TUNNEL_TOKEN（从 Cloudflare Zero Trust 获取）
#   2. 填写 BOT_TOKEN、ADMIN_IDS、WEBAPP_SESSION_SECRET
#   3. 填写 WEBHOOK_BASE_URL 和 SUBSCRIPTION_MINI_APP_URL
#      （必须与 Cloudflare Tunnel 中配置的公共主机名一致）

docker compose up -d
docker compose ps
```
