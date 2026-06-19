# Docker 部署

## 1. 准备环境

需要 Docker Engine 与 Docker Compose v2。生产环境建议使用 HTTPS 反向代理，分别暴露：

- 前端 Mini App：转发到 `127.0.0.1:8082`
- 后端 webhook/API：转发到 `127.0.0.1:8080`

## 2. 配置最小 `.env`

复制示例文件：

```bash
cp .env.example .env
```

至少填写：

- `BOT_TOKEN`
- `ADMIN_IDS`
- `WEBAPP_SESSION_SECRET`
- `POSTGRES_PASSWORD`

支付、套餐、Remnawave、外观、语言、汇率建议首次登录后台后配置。`.env` 中的 Remnawave 与支付字段只作为首次启动或后台未覆盖时的兜底值。

## 3. 启动

```bash
docker compose up -d --build
docker compose ps
```

迁移容器会先执行数据库迁移。容器和卷命名统一为 `remna-user-panel-*`。

## 4. 首次后台配置

打开前端域名并使用 Telegram 管理员账号进入后台，然后依次配置：

1. Settings：Webhook Base URL、默认货币、汇率源。
2. Tariffs：套餐目录，默认 USD。
3. Payments：EZPay 或 BEPUSDT。
4. Appearance：Logo、Favicon、主题。
5. Remnawave：面板 API、默认 traffic strategy、squads、HWID 设备限制。

## 5. 反向代理

后端 webhook 域名必须能访问：

- `/webhook/telegram` 或 Bot Token 形式的 Telegram webhook 路径
- `/webhook/panel`
- `/webhook/ezpay`
- `/webhook/bepusdt`
- `/healthz`

前端域名转发到 nginx 容器即可。
