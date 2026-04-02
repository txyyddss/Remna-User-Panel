# Remna User Panel

> VPN 拼车面板 + Telegram Bot — Go后端 + Vue.js前端

## ✨ 功能

- 🤖 **Telegram Mini App** — 完整的 Mini App 集成
- 💎 **TXB积分系统** — 签到、赌博、消费赠送、账单折扣
- 📡 **VPN订阅管理** — Remnawave 套餐购买/续费/换套餐
- 🎬 **Jellyfin影视** — 账户管理、Quick Connect、内容分级
- 💳 **双支付网关** — BEPusdt (USDT) + EZPay (支付宝/微信)
- 🌐 **线路切换** — 外部线路组一键切换
- 🔄 **IP更换** — 带冷却时间的IP更换
- 📊 **流量统计** — 按节点流量分布图
- ⚙️ **管理面板** — 配置管理、套餐管理

## 🚀 快速开始

### 后端

```bash
cd backend
cp config.example.json config.json
# 编辑 config.json 填入你的API密钥

# 构建
go build -o remna-panel ./cmd/server/

# 运行
./remna-panel
```

### 前端

```bash
cd frontend
npm install
npm run dev     # 开发模式
npm run build   # 生产构建
```

## ⚙️ 配置

复制 `backend/config.example.json` 到 `backend/config.json`，填入:

| 配置项 | 说明 |
|--------|------|
| `telegram.bot_token` | Telegram Bot Token |
| `telegram.admin_ids` | 管理员 Telegram ID 数组 |
| `remnawave.url/token` | Remnawave API 地址和Token |
| `jellyfin.url/token` | Jellyfin API 地址和Token |
| `bepusdt.*` | BEPusdt USDT支付网关配置 |
| `ezpay.*` | 易支付配置 |
| `ai.api_key` | DeepSeek API Key (群聊评分) |

支持热重载 — 修改 `config.json` 后自动生效。

## 📁 项目结构

```
backend/
├── cmd/server/main.go          # 入口
├── internal/
│   ├── config/                 # 配置管理
│   ├── database/               # SQLite + 迁移
│   ├── models/                 # 数据模型
│   ├── handlers/               # HTTP处理器
│   ├── middleware/              # 认证、响应
│   ├── services/               # 业务逻辑
│   ├── sdk/{remnawave,jellyfin,ezpay,bepusdt}/
│   ├── telegram/               # Bot
│   └── cron/                   # 定时任务

frontend/
├── src/
│   ├── api/                    # API客户端
│   ├── stores/                 # Pinia状态
│   ├── views/                  # 页面组件
│   └── styles/                 # 设计系统
```

## 🔧 CI/CD

- **后端**: Push to `main` → Go编译 → GitHub Release
- **前端**: Push to `main` (frontend/) → Vue构建 → `web` 分支

## 🚀 部署到生产环境

详细的生产部署步骤（前端 Cloudflare Pages + 后端 Debian 13 服务器）请参见 **[docs/deployment.md](docs/deployment.md)**。

## 📄 License

MIT
