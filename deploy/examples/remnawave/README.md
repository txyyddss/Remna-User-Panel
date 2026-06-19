# 与 Remnawave 面板共用 Docker Compose

详细文档见：[docs/getting-started/deployment-remnawave.md](../../../docs/getting-started/deployment-remnawave.md)

## 文件说明

- `docker-compose.yml`：Remna User Panel + Remnawave 面板共用 PostgreSQL/Redis 的完整部署
- `.env.example`：完整环境变量模板，包含两个项目的所有配置项

## 快速开始

```bash
cp .env.example .env
# 编辑 .env，至少填写：
#   - POSTGRES_PASSWORD
#   - PUBLIC_URL
#   - BOT_TOKEN、ADMIN_IDS
#   - WEBAPP_SESSION_SECRET、WEBHOOK_SECRET_TOKEN
#   - PANEL_API_URL、PANEL_API_KEY、PANEL_WEBHOOK_SECRET
#   - JWT_AUTH_SECRET、JWT_API_TOKENS_SECRET（Remnawave 面板）
#   - SUB_PUBLIC_DOMAIN（Remnawave 面板）

# 首次启动：db-init 会自动创建 remnawave 和 remna_user_panel 数据库
docker compose up -d
docker compose ps
```

## 变量命名说明

`.env` 中两个项目的变量自然共存，不存在冲突：

| 分类 | 变量示例 | 所属 |
|------|---------|------|
| 共享基础设施 | `POSTGRES_USER`, `POSTGRES_PASSWORD` | PostgreSQL 容器 |
| 共享基础设施 | `REMAWAVE_DB`, `USER_PANEL_DB` | 各自数据库名（db-init 自动创建） |
| 公共 URL | `PUBLIC_URL` | User Panel（webhook + Mini App） |
| User Panel | `BOT_TOKEN`, `ADMIN_IDS`, `PANEL_API_URL`, `WEBAPP_SESSION_SECRET`... | Remna User Panel |
| Remnawave 面板 | `JWT_AUTH_SECRET`, `SUB_PUBLIC_DOMAIN`, `TELEGRAM_BOT_TOKEN`... | Remnawave 面板 |
| 共享密钥 | `PANEL_WEBHOOK_SECRET` | 同时用于 User Panel 和 Remnawave 的 `WEBHOOK_SECRET_HEADER` |

> **注意**：默认语言、货币、Squad、通知天数等配置请在后台 Web UI → Settings 页面修改，无需设置环境变量。
