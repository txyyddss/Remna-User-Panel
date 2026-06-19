# Remnawave Minishop

## 开发

- [插件和扩展 API](development/plugins.md) — Python 插件钩子、领域事件、功能标志、迁移、本地化和管理面板板块扩展规则。

Remnawave Minishop — Telegram 机器人和 Mini App, 用于销售和管理 Remnawave 订阅。文档帮助部署技术栈、配置支付、套餐、管理面板、客服和公开控制台。

> 项目与 Remnawave Panel 配合工作: 面板存储用户和订阅, Minishop 负责 Telegram、支付、Mini App、套餐和运营管理。

## 核心功能

- **订阅销售** — 周期和流量套餐、流量补充、HWID 设备、[Premium Squads](features/tariffs.md#premium-сквады-и-отдельный-лимит)
- **用户生命周期** — 注册、试用期、续费、面板同步和流量警告.
- **Mini App** — 控制台、安装指南、[Telegram OAuth](features/telegram-auth.md)、[邮箱登录](features/email-login.md)和公开推荐链接.
- **运营工具** — 管理面板、客服工单、优惠码、群发、日志、[备份与恢复](features/backups.md)、`.env` 覆盖设置.
