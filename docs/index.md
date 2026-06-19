# Remna User Panel 中文文档

Remna User Panel 是面向 Telegram Mini App 的 Remnawave 用户面板。当前版本以 Go 后端、Svelte/Vite 前端、PostgreSQL 和 Redis 为基础，支付范围收敛为 EZPay 与 BEPUSDT。

## 文档目录

- [Docker 部署](getting-started/deployment.md)
- [与 Remnawave 面板共用部署](getting-started/deployment-remnawave.md)
- [Cloudflare Tunnel 部署](getting-started/deployment-cloudflared.md)
- [后台配置](configuration/settings.md)
- [支付配置](features/payments.md)
- [套餐与 plan_hash](features/tariffs.md)
- [安全建议](configuration/security.md)
- [维护、更新与备份](troubleshooting/maintenance.md)

## 核心约定

- `.env` 只用于最小启动：数据库、Redis、监听端口、Bot Token、管理员 ID、Session Secret、Webhook Secret、可信代理和日志级别。
- 套餐、Remnawave、支付、外观、语言、汇率等配置在管理后台维护。
- 前端下单只提交 `plan_hash` 与支付方式，金额由服务端按套餐快照、汇率和支付 provider 重新计算。
- 默认主币种是 USD，用户界面显示 USD 主价和 CNY/RMB 参考价。
- 后端单端口（8080）同时服务 webhook、Mini App 页面、API 和静态资源。
