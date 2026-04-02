# Usage & Configuration Guide

This documentation covers standard operational aspects of Remna User Panel once deployed.

## 1. Environment & Build

The application encompasses two main pipelines built specifically to target modern Web and CLI systems.

### Building Services
**Backend Context:**
```bash
cd backend
go mod tidy
go build -o remna-panel ./cmd/server
./remna-panel
```
By default, the backend expects a `config.json` running at port 8080 mapping to a SQLite database. If none exists, an example is generated at startup.

**Frontend Context:**
```bash
cd frontend
npm install
npm run build 
```
Output static files will sit at `frontend/dist/`. For server distributions, using Nginx or Caddy is ideal for serving HTML to standard `/api/v1` backend endpoints.

## 2. Setting Up Configuration Hooks (`config.json`)

The system configuration is mapped directly into `config.json` with hot-reload features supported:

*   **Telegram Bot Options**:
    Include `admin_ids` arrays and a valid token. The integrated bot exposes commands and triggers "daily check-in" routines allowing users to claim TXB credits.
*   **Dual-Payment Platforms**:
    You can freely adapt to exclusively crypto or fiat modes. 
    1. For `bepusdt`, add respective contract tokens and your callback IP structure.
    2. For `ezpay`, add the `apiid` and `apikey` to the provider structure.
*   **DeepSeek AI Integration**:
    Provides AI responses in bot chat queries, analyzing user context.

## 3. Telegram Mini-App Bootstrapping
Users entering the official Telegram bot can open the UI natively seamlessly without login pages, given proper valid headers are generated inside the TG native wrapper. On frontend build tools, configuring the host to match the actual Bot UI allows one-tap operations inside Telegram Apps alongside native OS browsers. 

## 4. IP Rotation & Constraints
The panel uniquely mitigates abusive crawling. Settings in IP rotation have explicit cooldown timers configured globally. Changes triggered inside the application dynamically force connection drops over Remnawave, refreshing routing sessions synchronously.
