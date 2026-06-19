# 与 Remnawave 面板共用 Docker Compose

详细文档见：[docs/getting-started/deployment-remnawave.md](../../../docs/getting-started/deployment-remnawave.md)

## 文件说明

- `docker-compose.yml`：Remna User Panel + Remnawave 面板共用 PostgreSQL/Redis 的完整部署
- `.env.example`：最小环境变量模板

## 快速开始

```bash
cp .env.example .env
# 编辑 .env 填写 BOT_TOKEN、ADMIN_IDS、密钥、Remnawave 面板地址

docker compose up -d postgres redis
docker compose exec postgres psql -U postgres -c "CREATE DATABASE remnawave;"
docker compose exec postgres psql -U postgres -c "CREATE DATABASE remna_user_panel;"
docker compose up -d
docker compose ps
```
