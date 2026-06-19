# 维护、更新与备份

## 更新

```bash
git pull
docker compose up -d --build
docker compose ps
```

## 备份

至少备份：

- PostgreSQL 数据库。
- `shop-data` 卷。
- `.env` 中的启动密钥。

示例：

```bash
docker compose exec postgres pg_dump -U "$POSTGRES_USER" "$POSTGRES_DB" > backup.sql
```

## 恢复

先停止服务，再恢复数据库和数据卷。恢复完成后运行：

```bash
docker compose up -d
docker compose logs -f backend
```

## 健康检查

- 后端：`/healthz`
- 前端：`/health`
- Docker：`docker compose ps`
