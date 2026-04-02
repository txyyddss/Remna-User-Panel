#!/bin/bash
set -euo pipefail

SERVICE_NAME="remna-backend"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"
APP_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_NAME="remna-backend"
BIN_PATH="${APP_DIR}/${BIN_NAME}"
INSTALLED_VERSION_FILE="${APP_DIR}/.installed-release"
GITHUB_REPO="txyyddss/Remna-User-Panel"
RELEASE_API_URL="https://api.github.com/repos/${GITHUB_REPO}/releases/latest"
RELEASE_ASSET_NAME="remna-panel-linux-amd64"
AUTOUPDATE_LOG="/var/log/remna-updater.log"
CRON_JOB="0 2 * * * cd ${APP_DIR} && ./manage.sh release-update >> ${AUTOUPDATE_LOG} 2>&1"

if [ "$EUID" -ne 0 ]; then
  echo "Please run as root (use sudo)"
  exit 1
fi

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required command: $1"
    exit 1
  fi
}

build_local_binary() {
  echo "Building ${BIN_NAME} from local source..."
  require_cmd go
  cd "$APP_DIR"
  go build -o "$BIN_NAME" ./cmd/server/
}

fetch_latest_release_info() {
  require_cmd curl
  require_cmd python3

  local json
  json="$(curl -fsSL "$RELEASE_API_URL")"

  LATEST_TAG="$(printf '%s' "$json" | python3 -c "import json,sys; print(json.load(sys.stdin)['tag_name'])")"
  LATEST_ASSET_URL="$(printf '%s' "$json" | python3 -c "import json,sys; data=json.load(sys.stdin); name='${RELEASE_ASSET_NAME}'; print(next((a['browser_download_url'] for a in data.get('assets', []) if a.get('name') == name), ''))")"

  if [ -z "${LATEST_TAG}" ] || [ -z "${LATEST_ASSET_URL}" ]; then
    echo "Failed to locate the latest release asset ${RELEASE_ASSET_NAME}"
    exit 1
  fi
}

download_release_binary() {
  fetch_latest_release_info

  local current_tag=""
  if [ -f "$INSTALLED_VERSION_FILE" ]; then
    current_tag="$(cat "$INSTALLED_VERSION_FILE")"
  fi

  if [ "${current_tag}" = "${LATEST_TAG}" ]; then
    echo "Already up to date at ${LATEST_TAG}."
    return 0
  fi

  echo "Downloading release ${LATEST_TAG}..."
  local tmp_file
  tmp_file="$(mktemp "${APP_DIR}/.${BIN_NAME}.XXXXXX")"
  curl -fsSL "$LATEST_ASSET_URL" -o "$tmp_file"
  chmod +x "$tmp_file"
  mv "$tmp_file" "$BIN_PATH"
  echo "$LATEST_TAG" > "$INSTALLED_VERSION_FILE"
  echo "Installed ${LATEST_TAG}."
}

install_service() {
  echo "Installing ${SERVICE_NAME}..."
  if [ ! -f "$BIN_PATH" ]; then
    if curl -fsSL "$RELEASE_API_URL" >/dev/null 2>&1; then
      download_release_binary
    else
      build_local_binary
    fi
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

  echo "Removing auto-updater if it exists..."
  disable_autoupdate
}

enable_autoupdate() {
  echo "Enabling release-aware auto-update (cron job)..."
  (crontab -l 2>/dev/null | grep -v "manage.sh release-update"; echo "$CRON_JOB") | crontab -
  echo "Auto-update enabled. New GitHub releases will be applied daily at 2:00 AM."
}

disable_autoupdate() {
  echo "Disabling auto-update..."
  crontab -l 2>/dev/null | grep -v "manage.sh release-update" | crontab - || true
  echo "Auto-update disabled."
}

restart_service() {
  echo "Restarting ${SERVICE_NAME}..."
  systemctl restart "${SERVICE_NAME}"
  echo "Service restarted."
}

update_service() {
  echo "Checking for a newer GitHub release..."
  download_release_binary
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
  echo "  update          Download the latest published release and restart"
  echo "  release-update  Alias of update; used by auto-update"
  echo "  build-local     Build the backend binary from the local checkout"
  echo "  autoupdate      Enable daily release-aware auto-update via cron"
  echo "  no-autoupdate   Disable the auto-updater cron job"
  echo ""
}

case "${1:-}" in
  install)
    install_service
    ;;
  uninstall)
    uninstall_service
    ;;
  restart)
    restart_service
    ;;
  update|release-update)
    update_service
    ;;
  build-local)
    build_local_binary
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
