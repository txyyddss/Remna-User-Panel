# Remna User Panel

Telegram Mini App 用户面板，用于销售和管理 Remnawave 订阅。后端使用 Go，前端使用 Svelte/Vite，数据层使用 PostgreSQL 与 Redis。

## 文档

中文文档入口：[`docs/index.md`](docs/index.md)

核心部署文档：[`docs/getting-started/deployment.md`](docs/getting-started/deployment.md)

## 镜像

- `ghcr.io/txyyddss/remna-user-panel-backend`
- `ghcr.io/txyyddss/remna-user-panel-worker`

Compose 容器、卷和二进制统一使用 `remna-user-panel-*` 命名。
