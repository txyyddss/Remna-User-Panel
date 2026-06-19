# Remna User Panel

Telegram Mini App 用户面板，用于销售和管理 Remnawave 订阅。后端使用 Go，前端使用 Svelte/Vite，数据层使用 PostgreSQL 与 Redis。

## 当前范围

- 用户端：订阅状态、套餐购买、支付、设备、客服、安装指南、语言切换。
- 管理端：用户、套餐、支付、设置、主题/Logo/Favicon、翻译、统计、日志、促销码、广告、群发、备份入口。
- 支付：仅保留 EZPay 与 BEPUSDT。
- 货币：默认 USD，前端显示 CNY/RMB 参考价。
- 下单：前端只提交 `plan_hash` 和支付方式，服务端按套餐快照计算金额。

## 快速启动

```bash
cp .env.example .env
nano .env
docker compose up -d --build
docker compose ps
```

`.env` 只需要填写最小启动项；Remnawave API、套餐、支付、汇率、外观和语言可以在首次登录后台后配置。`.env` 中的 Remnawave/支付字段会作为后台未覆盖时的兜底值。

## 文档

中文文档入口：[`docs/index.md`](docs/index.md)

核心部署文档：[`docs/getting-started/deployment.md`](docs/getting-started/deployment.md)

## 镜像

- `ghcr.io/<namespace>/remna-user-panel-backend`
- `ghcr.io/<namespace>/remna-user-panel-worker`

Compose 容器、卷和二进制统一使用 `remna-user-panel-*` 命名。
