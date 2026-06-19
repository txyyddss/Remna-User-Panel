# 支付配置

当前只保留两个支付 provider：

- EZPay：默认按 CNY 创建订单，服务端把 USD 套餐价按 USD/CNY 汇率换算为 CNY。
- BEPUSDT：按 USD fiat 金额创建 USDT 订单。

## 下单安全

前端不再提交金额、月份、流量或设备数量。支付请求格式：

```json
{
  "plan_hash": "plan_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "method": "ezpay:alipay",
  "renew_hwid_devices": true
}
```

后端按 `plan_hash` 从服务端套餐目录解析套餐，并保存：

- `base_amount` / `base_currency`
- `display_cny_amount`
- `fx_rate` / `fx_source` / `fx_updated_at`
- `plan_hash`
- `plan_snapshot`
- provider 实际收款 `amount` / `currency`

## Webhook

EZPay 使用 MD5 签名校验。BEPUSDT 使用回调签名校验。支付履约按订单状态幂等更新，重复 webhook 不会重复覆盖 `paid_at`。
