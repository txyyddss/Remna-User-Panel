# EZPay (MotionPay) API Documentation

> **Base URL**: Refer to your merchant dashboard for the actual API endpoint  
> **Request Format**: `application/x-www-form-urlencoded`  
> **Response Format**: JSON  
> **Signature Algorithm**: MD5  
> **Encoding**: UTF-8

## Table of Contents

1. [协议规则](#)
2. [页面跳转支付](#)
3. [API接口支付](#api)
4. [支付结果通知](#)
5. [MD5签名算法](#md5)
6. [支付方式列表](#)
7. [设备类型列表](#)
8. [[API]查询商户信息](#api)
9. [[API]查询结算记录](#api)
10. [[API]查询单个订单](#api)
11. [[API]批量查询订单](#api)
12. [[API]提交订单退款](#api)
13. [SDK下载](#sdk)
14. [常见面板对接教程/示例](#)

---

## 协议规则

URL地址：文档仅供参考，实际以商户后台接口地址为准

请求数据格式：application/x-www-form-urlencoded

返回数据格式：JSON

签名算法：MD5

字符编码：UTF-8


---

## 页面跳转支付

此接口可用于用户前台直接发起支付，使用form表单跳转或拼接成url跳转。

URL地址： https://motionpay.net/submit.php

请求方式： POST 或 GET（推荐POST，不容易被劫持或屏蔽）

请求参数说明：

| 字段名 | 变量名 | 必填 | 类型 | 示例值 | 描述 |
| --- | --- | --- | --- | --- | --- |
| 商户ID | pid | 是 | Int | 1001 |  |
| 支付方式 | type | 否 | String | alipay | 支付方式列表 |
| 商户订单号 | out_trade_no | 是 | String | 20230806151343349 |  |
| 异步通知地址 | notify_url | 是 | String | http://www.pay.com/notify_url.php | 服务器异步通知地址 |
| 跳转通知地址 | return_url | 是 | String | http://www.pay.com/return_url.php | 页面跳转通知地址 |
| 商品名称 | name | 是 | String | VIP会员 | 如超过127个字节会自动截取 |
| 商品金额 | money | 是 | String | 1.00 | 单位：元，最大2位小数 |
| 业务扩展参数 | param | 否 | String | 没有请留空 | 支付后原样返回 |
| 签名字符串 | sign | 是 | String | 202cb962ac59075b964b07152d234b70 | 签名算法 点此查看 |
| 签名类型 | sign_type | 是 | String | MD5 | 默认为MD5 |

支付方式（type）不传会跳转到收银台支付

### Usage Example

**HTML Form (POST)**:
```html
<form action="https://motionpay.net/submit.php" method="POST">
    <input type="hidden" name="pid" value="1001">
    <input type="hidden" name="type" value="alipay">
    <input type="hidden" name="out_trade_no" value="20230806151343349">
    <input type="hidden" name="notify_url" value="https://example.com/notify">
    <input type="hidden" name="return_url" value="https://example.com/return">
    <input type="hidden" name="name" value="VIP Membership">
    <input type="hidden" name="money" value="9.99">
    <input type="hidden" name="sign" value="COMPUTED_MD5_SIGN">
    <input type="hidden" name="sign_type" value="MD5">
    <button type="submit">Pay Now</button>
</form>
```

**URL Redirect (GET)**:
```
https://motionpay.net/submit.php?pid=1001&type=alipay&out_trade_no=20230806151343349&notify_url=https://example.com/notify&return_url=https://example.com/return&name=VIP&money=9.99&sign=COMPUTED_MD5_SIGN&sign_type=MD5
```


---

## API接口支付

此接口可用于服务器后端发起支付请求，会返回支付二维码链接或支付跳转url。

URL地址： https://motionpay.net/mapi.php

请求方式： POST

请求参数说明：

| 字段名 | 变量名 | 必填 | 类型 | 示例值 | 描述 |
| --- | --- | --- | --- | --- | --- |
| 商户ID | pid | 是 | Int | 1001 |  |
| 支付方式 | type | 是 | String | alipay | 支付方式列表 如需收银台，传: cashier |
| 商户订单号 | out_trade_no | 是 | String | 20230806151343349 |  |
| 异步通知地址 | notify_url | 是 | String | http://www.pay.com/notify_url.php | 服务器异步通知地址 |
| 跳转通知地址 | return_url | 否 | String | http://www.pay.com/return_url.php | 页面跳转通知地址 |
| 商品名称 | name | 是 | String | VIP会员 | 如超过127个字节会自动截取 |
| 商品金额 | money | 是 | String | 1.00 | 单位：元，最大2位小数 |
| 用户IP地址 | clientip | 是 | String | 192.168.1.100 | 用户发起支付的IP地址 |
| 设备类型 | device | 否 | String | pc 建议按实际传入，适配最佳体验 | 根据当前用户浏览器的UA判断， 传入用户所使用的浏览器 或设备类型，默认为pc 设备类型列表 |
| 获取原始链接 | rawurl | 否 | Int | 1 默认0，返回兼容性好的收款链接payurl 传1时，将返回二维码链接qrcode | 默认0，返回兼容性好的收款链接payurl 传1时，将返回二维码链接qrcode |
| 业务扩展参数 | param | 否 | String | 没有请留空 | 支付后原样返回 |
| 签名字符串 | sign | 是 | String | 202cb962ac59075b964b07152d234b70 | 签名算法 点此查看 |
| 签名类型 | sign_type | 是 | String | MD5 | 默认为MD5 |

返回结果（json）：

| 字段名 | 变量名 | 类型 | 示例值 | 描述 |
| --- | --- | --- | --- | --- |
| 返回状态码 | code | Int | 1 | 1为成功，其它值为失败 |
| 返回信息 | msg | String |  | 失败时返回原因 |
| 订单号 | trade_no | String | 20230806151343349 | 支付订单号 |
| 支付跳转url | payurl | String | https://motionpay.net/pay/wxpay/202010903/ | 如果返回该字段，则直接跳转到该url支付 |
| 二维码链接 | qrcode | String | weixin://wxpay/bizpayurl?pr=04IPMKM | 如果返回该字段，则根据该url生成二维码 |
| 小程序跳转url | urlscheme | String | weixin://dl/business/?ticket=xxx | 如果返回该字段，则使用js跳转该url，可发起微信小程序支付 |

注：payurl、qrcode、urlscheme 不一定都返回，建议都适配

### Usage Example

```bash
curl -X POST https://motionpay.net/mapi.php \
  -d "pid=1001" \
  -d "type=alipay" \
  -d "out_trade_no=20230806151343349" \
  -d "notify_url=https://example.com/notify" \
  -d "name=VIP Membership" \
  -d "money=9.99" \
  -d "clientip=192.168.1.100" \
  -d "sign=COMPUTED_MD5_SIGN" \
  -d "sign_type=MD5"
```

**Success Response**:
```json
{
    "code": 1,
    "msg": "",
    "trade_no": "20230806151343349021",
    "payurl": "https://motionpay.net/pay/wxpay/202010903/",
    "qrcode": "weixin://wxpay/bizpayurl?pr=04IPMKM",
    "urlscheme": "weixin://dl/business/?ticket=xxx"
}
```

**Error Response**:
```json
{
    "code": -1,
    "msg": "Invalid signature"
}
```


---

## 支付结果通知

通知类型：服务器异步通知（notify_url）、页面跳转通知（return_url）

请求方式：GET

请求参数说明：

| 字段名 | 变量名 | 必填 | 类型 | 示例值 | 描述 |
| --- | --- | --- | --- | --- | --- |
| 商户ID | pid | 是 | Int | 1001 |  |
| 易支付订单号 | trade_no | 是 | String | 20230806151343349021 | 魔神支付订单号 |
| 商户订单号 | out_trade_no | 是 | String | 20230806151343349 | 商户系统内部的订单号 |
| 支付方式 | type | 是 | String | alipay | 支付方式列表 |
| 商品名称 | name | 是 | String | VIP会员 |  |
| 商品金额 | money | 是 | String | 1.00 |  |
| 支付状态 | trade_status | 是 | String | TRADE_SUCCESS | 只有TRADE_SUCCESS是成功 |
| 业务扩展参数 | param | 否 | String |  |  |
| 签名字符串 | sign | 是 | String | 202cb962ac59075b964b07152d234b70 | 签名算法 点此查看 |
| 签名类型 | sign_type | 是 | String | MD5 | 默认为MD5 |

收到异步通知后，需返回success以表示服务器接收到了订单通知

### Usage Example

When a payment is completed, the system sends a GET request to your `notify_url`:

```
GET https://example.com/notify?pid=1001&trade_no=20230806151343349021&out_trade_no=20230806151343349&type=alipay&name=VIP&money=1.00&trade_status=TRADE_SUCCESS&sign=xxx&sign_type=MD5
```

**Your server must respond with the plain text**: `success`

```python
# Python (Flask) example
@app.route('/notify')
def payment_notify():
    # 1. Verify the sign parameter
    # 2. Check trade_status == "TRADE_SUCCESS"
    # 3. Process the order in your database
    return "success"
```


---

## MD5签名算法

1、将发送或接收到的所有参数按照参数名ASCII码从小到大排序（a-z），sign、sign_type、值为空或0时，不参与签名

2、将排序后的参数拼接成URL键值对的格式，例如 a=b&c=d&e=f ，参数值不要进行url编码。

3、再将拼接好的字符串与商户密钥KEY进行MD5加密得出sign签名参数， sign = md5 ( a=b&c=d&e=f + KEY ) （注意：+ 为各语言的拼接符，不是字符！），md5结果为小写。

4、具体签名与发起支付的示例代码可下载SDK查看。


---

## 支付方式列表

| 调用值 | 描述 |
| --- | --- |
| alipay | 支付宝 |
| wxpay | 微信支付 |
| usdt | USDT |


---

## 设备类型列表

| 调用值 | 描述 |
| --- | --- |
| pc | 电脑浏览器 |
| mobile | 手机浏览器 |
| qq | 手机QQ内浏览器 |
| wechat | 微信内浏览器 |
| alipay | 支付宝客户端 |


---

## [API]查询商户信息

URL地址： https://motionpay.net/api.php?act=query&pid={商户ID}&sign={32位签名字符串}

请求参数说明：

| 字段名 | 变量名 | 必填 | 类型 | 示例值 | 描述 |
| --- | --- | --- | --- | --- | --- |
| 操作类型 | act | 是 | String | query | 此API固定值 |
| 商户ID | pid | 是 | Int | 1001 |  |
| 签名字符串 | sign | 是 | String | 202cb962ac59075b964b07152d234b70 | 签名算法 点此查看 |
| 签名类型 | sign_type | 否 | String | MD5 | 默认为MD5 |

返回结果：

| 字段名 | 变量名 | 类型 | 示例值 | 描述 |
| --- | --- | --- | --- | --- |
| 返回状态码 | code | Int | 1 | 1为成功，其它值为失败 |
| 商户ID | pid | Int | 1001 | 商户ID |
| 商户密钥 | key | String(32) | 89unJUB8HZ54Hj7x4nUj56HN4nUzUJ8i | 商户密钥 |
| 商户状态 | active | Int | 1 | 1为正常，0为封禁 |
| 商户余额 | money | String | 0.00 | 商户所拥有的余额 |
| 结算方式 | type | Int | 1 | 1:支付宝,2:微信,3:QQ,4:银行卡 |
| 结算地址 | account | String | admin@pay.com | 结算的支付宝账号 |
| 结算姓名 | username | String | 张三 | 结算的支付宝姓名 |
| 订单总数 | orders | Int | 30 | 订单总数统计 |
| 今日订单 | order_today | Int | 15 | 今日订单数量 |
| 昨日订单 | order_lastday | Int | 15 | 昨日订单数量 |
| 今日流水 | amount_today | Float | 3510.12 | 今日流水（扣除手续费后） |
| 昨日流水 | amount_yesterday | Float | 4233.42 | 昨日流水（扣除手续费后） |

### Usage Example

```bash
curl "https://motionpay.net/api.php?act=query&pid=1001&sign=COMPUTED_MD5_SIGN&sign_type=MD5"
```


---

## [API]查询结算记录

URL地址： https://motionpay.net/api.php?act=settle&pid={商户ID}&sign={32位签名字符串}

请求参数说明：

| 字段名 | 变量名 | 必填 | 类型 | 示例值 | 描述 |
| --- | --- | --- | --- | --- | --- |
| 操作类型 | act | 是 | String | settle | 此API固定值 |
| 商户ID | pid | 是 | Int | 1001 |  |
| 签名字符串 | sign | 是 | String | 202cb962ac59075b964b07152d234b70 | 签名算法 点此查看 |
| 签名类型 | sign_type | 是 | String | MD5 | 默认为MD5 |

返回结果：

| 字段名 | 变量名 | 类型 | 示例值 | 描述 |
| --- | --- | --- | --- | --- |
| 返回状态码 | code | Int | 1 | 1为成功，其它值为失败 |
| 返回信息 | msg | String | 查询结算记录成功！ |  |
| 结算记录 | data | Array | 结算记录列表 |  |

### Usage Example

```bash
curl "https://motionpay.net/api.php?act=settle&pid=1001&sign=COMPUTED_MD5_SIGN&sign_type=MD5"
```


---

## [API]查询单个订单

URL地址： https://motionpay.net/api.php?act=order&pid={商户ID}&out_trade_no={商户订单号}&sign={32位签名字符串}

请求参数说明：

| 字段名 | 变量名 | 必填 | 类型 | 示例值 | 描述 |
| --- | --- | --- | --- | --- | --- |
| 操作类型 | act | 是 | String | order | 此API固定值 |
| 商户ID | pid | 是 | Int | 1001 |  |
| 系统订单号 | trade_no | 选择 | String | 20230806151343312 |  |
| 商户订单号 | out_trade_no | 选择 | String | 20230806151343349 |  |
| 签名字符串 | sign | 是 | String | 202cb962a..7152d234b70 | 签名算法 点此查看 |
| 签名类型 | sign_type | 是 | String | MD5 | 默认为MD5 |

提示：系统订单号 和 商户订单号 二选一传入即可，如果都传入以系统订单号为准！

返回结果：

| 字段名 | 变量名 | 类型 | 示例值 | 描述 |
| --- | --- | --- | --- | --- |
| 返回状态码 | code | Int | 1 | 1为成功，其它值为失败 |
| 返回信息 | msg | String | 查询订单号成功！ |  |
| 易支付订单号 | trade_no | String | 2023080622555342651 | 魔神支付订单号 |
| 商户订单号 | out_trade_no | String | 20230806151343349 | 商户系统内部的订单号 |
| 第三方订单号 | api_trade_no | String | 20230806151343349 | 支付宝微信等接口方订单号 |
| 支付方式 | type | String | alipay | 支付方式列表 |
| 商户ID | pid | Int | 1001 | 发起支付的商户ID |
| 创建订单时间 | addtime | String | 2023-08-06 22:55:52 |  |
| 完成交易时间 | endtime | String | 2023-08-06 22:55:52 |  |
| 商品名称 | name | String | VIP会员 |  |
| 商品金额 | money | String | 1.00 |  |
| 支付状态 | status | Int | 0 | 1为支付成功，0为未支付 |
| 业务扩展参数 | param | String |  | 默认留空 |
| 支付者账号 | buyer | String |  | 默认留空 |

### Usage Example

```bash
curl "https://motionpay.net/api.php?act=order&pid=1001&sign=COMPUTED_MD5_SIGN&sign_type=MD5"
```


---

## [API]批量查询订单

URL地址： https://motionpay.net/api.php?act=orders&pid={商户ID}&sign={32位签名字符串}

请求参数说明：

| 字段名 | 变量名 | 必填 | 类型 | 示例值 | 描述 |
| --- | --- | --- | --- | --- | --- |
| 操作类型 | act | 是 | String | orders | 此API固定值 |
| 商户ID | pid | 是 | Int | 1001 |  |
| 查询数量 | limit | 否 | Int | 20 | 返回的订单数量，最大50 |
| 起始位置 | offset | 否 | Int | 0 | 当前查询的起始位，最大2000 |
| 签名字符串 | sign | 是 | String | 202cb962a..7152d234b70 | 签名算法 点此查看 |
| 签名类型 | sign_type | 是 | String | MD5 | 默认为MD5 |

返回结果：

| 字段名 | 变量名 | 类型 | 示例值 | 描述 |
| --- | --- | --- | --- | --- |
| 返回状态码 | code | Int | 1 | 1为成功，其它值为失败 |
| 返回信息 | msg | String | 查询结算记录成功！ |  |
| 订单列表 | data | Array |  | 订单列表 |

### Usage Example

```bash
curl "https://motionpay.net/api.php?act=orders&pid=1001&sign=COMPUTED_MD5_SIGN&sign_type=MD5"
```


---

## [API]提交订单退款

需要先在商户后台开启订单退款API接口开关，才能调用该接口发起订单退款

URL地址： https://motionpay.net/api.php?act=refund

请求方式： POST

请求参数说明：

| 字段名 | 变量名 | 必填 | 类型 | 示例值 | 描述 |
| --- | --- | --- | --- | --- | --- |
| 商户ID | pid | 是 | Int | 1001 |  |
| 易支付订单号 | trade_no | 特殊可选 | String | 20230806151343349021 | 易支付订单号 |
| 商户订单号 | out_trade_no | 特殊可选 | String | 20230806151343349 | 订单支付时传入的商户订单号，商家自定义且保证商家系统中唯一 |
| 退款金额 | money | 是 | String | 1.50 | 少数通道需要与原订单金额一致 |
| 签名字符串 | sign | 是 | String | 202cb962a..7152d234b70 | 签名算法 点此查看 |
| 签名类型 | sign_type | 是 | String | MD5 | 默认为MD5 |

注：trade_no、out_trade_no 不能同时为空，如果都传了以trade_no为准

返回结果：

| 字段名 | 变量名 | 类型 | 示例值 | 描述 |
| --- | --- | --- | --- | --- |
| 返回状态码 | code | Int | 1 | 1为成功，其它值为失败 |
| 返回信息 | msg | String | 退款成功 |  |

### Usage Example

```bash
curl -X POST https://motionpay.net/api.php?act=refund \
  -d "pid=1001" \
  -d "trade_no=20230806151343349021" \
  -d "money=1.50" \
  -d "sign=COMPUTED_MD5_SIGN" \
  -d "sign_type=MD5"
```


---

