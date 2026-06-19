# Docker Compose 示例

权威的部署方案文档位于 [docs/getting-started/deployment.md](../../docs/getting-started/deployment.md)。

此文件夹仅包含可用的 compose 示例和配置。详细说明不在此重复, 以便文档网站和 README 导航使用同一来源。

应用文件 (`/app/data`: 套餐、主题、Logo) 从所选 `docker-compose.yml` 旁的 `data` 文件夹挂载到 `migrate`、`backend` 和 `worker`。自定义主题请创建 `data/themes`; 手动套餐目录请使用 `data/tariffs.json`。

| 文件夹   | 文档                                                                                                                                           |
| ------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `caddy`    | [使用 Caddy 部署](../../docs/getting-started/deployment.md)                                       |
| `nginx`    | [使用 Nginx 部署](../../docs/getting-started/deployment.md)                                                                                 |
| `cloudflared` | [使用 Cloudflare Tunnel 部署](../../docs/getting-started/deployment-cloudflared.md)                                                      |
| `remnawave` | [与 Remnawave 面板共用部署](../../docs/getting-started/deployment-remnawave.md)                                                      |
| `newt`     | [通过 Pangolin / Newt 部署](../../docs/getting-started/deployment.md)                                                      |
| `no-proxy` | [无反向代理启动](../../docs/getting-started/deployment.md)                                |
| `mail`     | [部署本地 SMTP 服务器](../../docs/features/email-login.md) |
