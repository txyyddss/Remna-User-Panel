# 安全建议

## 必填密钥

生产环境必须设置 `WEBAPP_SESSION_SECRET`。系统不会再使用固定开发密钥回退。

## 请求安全

后端启用：

- 严格 JSON 解码和请求体大小限制。
- CSRF 校验。
- 安全响应头。
- 静态资源路径穿越防护。
- 上传 MIME 与尺寸检查。
- 支付 webhook 签名校验。

## 反向代理

只把可信代理写入 `TRUSTED_PROXIES`。不要直接信任公网来源的 `X-Forwarded-For`。

使用 Cloudflare Tunnel 时，cloudflared 容器在 Docker 网络内运行，默认的 `TRUSTED_PROXIES`（包含 `10.0.0.0/8`、`172.16.0.0/12`、`192.168.0.0/16`）已经覆盖。无需额外添加 Cloudflare 的公网 IP 段。

## 支付

支付金额只由后端计算。任何来自前端的金额、月份、流量、设备数都不作为下单依据。
