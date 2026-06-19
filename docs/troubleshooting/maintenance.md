# 维护

计划维护通常包括更新镜像、检查迁移、日志和备份。自动 ZIP 备份和恢复的详细说明见[备份与恢复](../features/backups.md)。

## 更新

```bash
docker compose pull
docker compose up -d
docker compose logs -f migrate backend worker
```

## PostgreSQL 备份

```bash
docker compose exec -T postgres sh -c 'pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB"' > backup.sql
```

## 自动备份

Worker 可以收集包含 PostgreSQL 转储和 Compose 文件夹快照的 ZIP 归档, 发送到 Telegram 并在 `data/backups` 中保留最近归档。配置和恢复见[独立章节](../features/backups.md)。

更改 `.env` 中的备份设置后, 重启 backend 和 worker:

```bash
docker compose up -d --build backend worker
docker compose logs -f backend worker
```

从归档恢复可在管理面板 **系统 -> 备份** 中操作。也可手动上传 ZIP, 选择 `数据库` 和/或 `Compose 文件夹`, backend 会在启动前验证归档。详情: [备份与恢复](../features/backups.md#восстановление-из-админки)。

## 事后检查

- `docker compose ps`
- `docker compose logs -f backend worker frontend`
- 访问 backend 域名的 `/healthz`
- 登录 Mini App 和管理面板
- 测试支付或测试激活
