# 套餐与 `plan_hash`

套餐目录默认使用 USD：

```json
{
  "default_currency": "usd",
  "tariffs": []
}
```

每个可购买选项会生成稳定 `plan_hash`。签名字段包括套餐 key、销售模式、周期、流量、价格和币种。字段不变时 hash 不变；价格、币种、数量或模式变化时 hash 自动变化。

## 前端显示

接口返回 USD 主价格和 CNY 参考价：

```json
{
  "plan_hash": "plan_xxx",
  "base_amount": 9,
  "base_currency": "USD",
  "display_cny_amount": 64.8,
  "fx_rate": 7.2,
  "fx_source": "frankfurter"
}
```

前端展示为 `USD 主价 + 约 RMB/CNY`。
