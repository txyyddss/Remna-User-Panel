#!/bin/bash
# manage.sh - Automates installation, uninstallation, and updating of remna-backend Systemd service
set -e

SERVICE_NAME="remna-backend"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"
APP_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_NAME="remna-backend"
BIN_PATH="${APP_DIR}/${BIN_NAME}"
CRON_JOB="0 2 * * * cd ${APP_DIR} && git pull origin main && go build -o ${BIN_NAME} cmd/server/main.go && systemctl restart ${SERVICE_NAME} >> /var/log/remna-updater.log 2>&1"

# Check root
if [ "$EUID" -ne 0 ]; then
  echo "Please run as root (use sudo)"
  exit 1
fi

install_service() {
    echo "Installing ${SERVICE_NAME}..."
    if [ ! -f "$BIN_PATH" ]; then
        echo "Binary not found at ${BIN_PATH}. Compiling..."
        cd "$APP_DIR"
        go build -o "$BIN_NAME" cmd/server/main.go
    fi

    cat > "$SERVICE_FILE" <<EOF
[Unit]
Description=Remna User Panel Backend Service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=${APP_DIR}
ExecStart=${BIN_PATH}
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    systemctl enable "${SERVICE_NAME}"
    systemctl start "${SERVICE_NAME}"
    echo "Service installed and started successfully."
}

uninstall_service() {
    echo "Uninstalling ${SERVICE_NAME}..."
    systemctl stop "${SERVICE_NAME}" || true
    systemctl disable "${SERVICE_NAME}" || true
    rm -f "$SERVICE_FILE"
    systemctl daemon-reload
    echo "Service uninstalled successfully."
    
    echo "Removing auto-updater if exists..."
    disable_autoupdate
}

enable_autoupdate() {
    echo "Enabling auto-updater (cron job)..."
    (crontab -l 2>/dev/null | grep -v "git pull origin main.*${SERVICE_NAME}"; echo "$CRON_JOB") | crontab -
    echo "Auto-updater enabled. Will run daily at 2:00 AM."
}

disable_autoupdate() {
    echo "Disabling auto-updater..."
    crontab -l 2>/dev/null | grep -v "git pull origin main.*${SERVICE_NAME}" | crontab - || true
    echo "Auto-updater disabled."
}

restart_service() {
    echo "Restarting ${SERVICE_NAME}..."
    systemctl restart "${SERVICE_NAME}"
    echo "Service restarted."
}

update_service() {
    echo "Updating ${SERVICE_NAME} manually..."
    cd "$APP_DIR"
    git pull origin main
    go build -o "$BIN_NAME" cmd/server/main.go
    restart_service
    echo "Update complete."
}

show_help() {
    echo "Usage: ./manage.sh [command]"
    echo ""
    echo "Commands:"
    echo "  install         Install and start the systemd service"
    echo "  uninstall       Stop and remove the systemd service"
    echo "  restart         Restart the service"
    echo "  update          Pull from git, compile and restart"
    echo "  autoupdate      Enable daily auto-update via cron"
    echo "  no-autoupdate   Disable auto-updater cron job"
    echo ""
}

case "$1" in
    install)
        install_service
        ;;
    uninstall)
        uninstall_service
        ;;
    restart)
        restart_service
        ;;
    update)
        update_service
        ;;
    autoupdate)
        enable_autoupdate
        ;;
    no-autoupdate)
        disable_autoupdate
        ;;
    *)
        show_help
        exit 1
        ;;
esac
