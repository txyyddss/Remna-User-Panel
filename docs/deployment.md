# Deployment Guide

## Frontend — Cloudflare Pages

### Prerequisites

- A [Cloudflare](https://cloudflare.com) account
- The repository hosted on GitHub (already the case for this project)

### 1. Build settings

When creating a new Cloudflare Pages project, use the following settings:

| Setting | Value |
|---|---|
| **Framework preset** | None |
| **Build command** | `npm run build` |
| **Build output directory** | `dist` |
| **Root directory** | `frontend` |
| **Node.js version** | `20` (set in Environment Variables as `NODE_VERSION=20`) |

### 2. Environment variables

Set the following environment variable in **Settings → Environment variables**:

| Variable | Value | Notes |
|---|---|---|
| `VITE_API_BASE_URL` | `https://api.your-domain.com` | Your backend public URL |

> [!IMPORTANT]
> All frontend env variables must be prefixed with `VITE_` to be exposed to the browser by Vite.

### 3. Connect via Cloudflare Pages dashboard

1. Go to **Workers & Pages → Create → Pages → Connect to Git**
2. Select your GitHub repository and click **Begin setup**
3. Fill in the build settings from the table above
4. Add the environment variable `VITE_API_BASE_URL`
5. Click **Save and Deploy**

Cloudflare Pages will automatically redeploy on every push to `main`.

### 4. Custom domain (optional)

In **Pages project → Custom domains**, click **Set up a custom domain** and follow the prompts. Cloudflare will provision a free TLS certificate automatically.

### 5. SPA routing fix

Create the file `frontend/public/_redirects` with the following content so client-side routing works correctly:

```
/*    /index.html   200
```

---

## Backend — Debian 13 Server

### Prerequisites

- A Debian 13 (Trixie) VPS/dedicated server with root or `sudo` access
- Go 1.22+ installed ([official instructions](https://go.dev/doc/install))
- `git` installed (`apt install git`)
- A domain name pointed at the server (recommended for TLS)

### 1. Install Go

```bash
# Download and install Go (replace version as needed)
wget https://go.dev/dl/go1.24.2.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.24.2.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
go version
```

### 2. Clone the repository and configure

```bash
git clone https://github.com/<your-org>/Remna-User-Panel.git /opt/remna
cd /opt/remna/backend

cp config.example.json config.json
nano config.json   # fill in all required values (see table below)
```

**Key config values to fill in:**

| Key | Description |
|---|---|
| `server.host` | `0.0.0.0` (listen on all interfaces) |
| `server.port` | `8080` (or your preferred port) |
| `server.api_secret` | A long, random secret string |
| `telegram.bot_token` | Your Telegram Bot token |
| `telegram.admin_ids` | Array of admin Telegram user IDs |
| `telegram.webhook_url` | `https://api.your-domain.com/telegram/webhook` |
| `remnawave.url` / `.token` | Remnawave API base URL and token |
| `jellyfin.url` / `.token` | Jellyfin API base URL and token |
| `bepusdt.*` | BEPusdt payment gateway credentials |
| `ezpay.*` | EZPay (Alipay/WeChat) credentials |
| `ai.api_key` | DeepSeek API key (if AI features enabled) |

### 3. Build the binary

```bash
cd /opt/remna/backend
go build -o remna-backend ./cmd/server/
```

### 4. Install as a systemd service

Use the included `manage.sh` script:

```bash
cd /opt/remna/backend
sudo chmod +x manage.sh
sudo ./manage.sh install
```

This compiles the binary (if not already done), installs a systemd service named `remna-backend`, enables it on boot, and starts it immediately.

**Useful service commands:**

```bash
sudo systemctl status remna-backend   # view status
sudo journalctl -u remna-backend -f   # follow logs
sudo ./manage.sh restart              # restart
sudo ./manage.sh update               # git pull + rebuild + restart
sudo ./manage.sh autoupdate           # enable daily auto-update at 02:00
```

### 5. Expose the backend with Nginx (reverse proxy + TLS)

Install Nginx and Certbot:

```bash
sudo apt install -y nginx python3-certbot-nginx
```

Create `/etc/nginx/sites-available/remna`:

```nginx
server {
    listen 80;
    server_name api.your-domain.com;

    location / {
        proxy_pass         http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header   Host              $host;
        proxy_set_header   X-Real-IP         $remote_addr;
        proxy_set_header   X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto $scheme;
        proxy_read_timeout 120s;
    }
}
```

Enable the site and obtain a certificate:

```bash
sudo ln -s /etc/nginx/sites-available/remna /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl reload nginx
sudo certbot --nginx -d api.your-domain.com
```

Certbot will automatically configure HTTPS and set up a renewal cron job.

### 6. Firewall

Allow only HTTP, HTTPS, and SSH:

```bash
sudo apt install -y ufw
sudo ufw allow OpenSSH
sudo ufw allow 'Nginx Full'
sudo ufw enable
```

Do **not** expose port `8080` directly — all traffic should go through Nginx.

### 7. Register the Telegram webhook

After the service is running and TLS is configured, register the webhook with Telegram:

```bash
curl "https://api.telegram.org/bot<YOUR_BOT_TOKEN>/setWebhook?url=https://api.your-domain.com/telegram/webhook"
```

You should receive `{"ok":true,"result":true}`.

### 8. Updates

To update the backend manually:

```bash
cd /opt/remna/backend
sudo ./manage.sh update
```

Or enable unattended nightly updates:

```bash
sudo ./manage.sh autoupdate
```

---

## Architecture overview

```
User (Telegram Mini App)
        │
        ▼
Cloudflare Pages          (frontend static files, CDN, free TLS)
        │  VITE_API_BASE_URL
        ▼
Nginx (TLS termination, reverse proxy)
        │
        ▼
remna-backend :8080       (Go binary, systemd service)
        │
        ├── Remnawave API
        ├── Jellyfin API
        ├── BEPusdt / EZPay
        └── SQLite database (./data.db)
```
