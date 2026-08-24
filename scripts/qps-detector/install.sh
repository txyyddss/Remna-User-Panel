#!/usr/bin/env bash
set -euo pipefail

config_path=/etc/tx-carpool-qps-detector.conf
state_path=/var/lib/tx-carpool-qps-detector/last-report
reporter_path=/usr/local/sbin/tx-carpool-qps-detector-report
cron_path=/etc/cron.d/tx-carpool-qps-detector

require_root() { [ "${EUID}" -eq 0 ] || { printf '%s\n' 'Run as root.' >&2; exit 1; }; }
remove() { rm -f "$cron_path" "$reporter_path" "$config_path"; rm -rf /var/lib/tx-carpool-qps-detector; printf '%s\n' 'QPS detector reporter removed.'; }
require_debian() { [ -r /etc/debian_version ] && grep -q '^13' /etc/debian_version || { printf '%s\n' 'Debian 13 is required.' >&2; exit 1; }; command -v docker >/dev/null; command -v curl >/dev/null; command -v flock >/dev/null; systemctl is-active --quiet cron || { printf '%s\n' 'The cron service must be active.' >&2; exit 1; }; }
install_reporter() {
  install -d -m 700 /var/lib/tx-carpool-qps-detector
  install -m 600 /dev/null "$config_path"
  printf 'API_URL=%q\nNODE_TOKEN=%q\nINTERVAL_MINUTES=%q\n' "$api_url" "$node_token" "$interval" > "$config_path"
  cat > "$reporter_path" <<'REPORTER'
#!/usr/bin/env bash
set -euo pipefail
config_path=/etc/tx-carpool-qps-detector.conf
state_path=/var/lib/tx-carpool-qps-detector/last-report
lock_path=/var/lib/tx-carpool-qps-detector/lock
umask 077
exec 9>"$lock_path"
flock -n 9 || exit 0
source "$config_path"
now=$(date +%s)
last=0
[ -r "$state_path" ] && last=$(cat "$state_path")
[ "$now" -ge "$((last + INTERVAL_MINUTES * 60))" ] || exit 0
payload=$(mktemp)
trap 'rm -f "$payload"' EXIT
docker exec remnanode xlogs > "$payload"
curl --fail --silent --show-error --config <(printf '%s\n' "header = Authorization: Bearer $NODE_TOKEN" "request = POST" "data-binary = @$payload" "url = $API_URL/api/v1/agents/qps-reports") > /dev/null
printf '%s\n' "$now" > "$state_path"
REPORTER
  chmod 700 "$reporter_path"
  printf '* * * * * root %s\n' "$reporter_path" > "$cron_path"
  chmod 600 "$cron_path"
}
main() {
  require_root
  if [ "${1:-install}" = "uninstall" ]; then remove; return; fi
  require_debian
  read -r -p 'API domain (https://panel.example): ' api_url
  read -r -s -p 'Node token: ' node_token; printf '\n'
  read -r -p 'Report interval in minutes [30]: ' interval
  interval=${interval:-30}
  [[ "$api_url" =~ ^https?://[^[:space:]]+$ ]] && [[ "$node_token" =~ ^[A-Za-z0-9_-]{20,}$ ]] && [[ "$interval" =~ ^[1-9][0-9]{0,3}$ ]] || { printf '%s\n' 'Invalid configuration.' >&2; exit 1; }
  install_reporter
  printf '%s\n' 'QPS detector reporter installed.'
}
main "$@"
