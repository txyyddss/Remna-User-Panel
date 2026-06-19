# 后台配置

管理后台是运行时配置的唯一权威来源。`.env` 只保留启动所需密钥和基础设施连接参数。

## 支付与回调

在 Settings 中填写 `WEBHOOK_BASE_URL`，格式为后端公网 HTTPS 地址，例如：

```text
https://api.example.com
```

系统会生成：

- EZPay：`https://api.example.com/webhook/ezpay`
- BEPUSDT：`https://api.example.com/webhook/bepusdt`

## 汇率

默认 provider 是 `frankfurter`。也可以选择：

- `exchange_rate_api`
- `custom`

汇率会缓存，provider 失败时使用最后一次有效缓存；没有缓存时使用保守 fallback，并在后台状态中标记为 stale。

## 语言与外观

语言文件位于 `locales/zh.json` 与 `locales/en.json`。后台翻译编辑器写入覆盖配置，不需要改 `.env`。

Logo、Favicon 和主题都在 Appearance 中维护。
