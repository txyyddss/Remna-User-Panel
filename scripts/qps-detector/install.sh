#!/usr/bin/env bash
set -euo pipefail

config_path=/etc/tx-carpool-qps-detector.conf
reporter_path=/usr/local/sbin/tx-carpool-qps-detector-report
cron_path=/etc/cron.d/tx-carpool-qps-detector

require_root() { [ "${EUID}" -eq 0 ] || { printf '%s\n' 'Run as root.' >&2; exit 1; }; }
remove() { rm -f "$cron_path" "$reporter_path" "$config_path"; rm -rf /var/lib/tx-carpool-qps-detector; printf '%s\n' 'QPS detector reporter removed.'; }
print_logs() {
  command -v docker >/dev/null || { printf '%s\n' 'Docker is required.' >&2; exit 1; }
  docker exec remnanode sh -c '
    [ -r "$1" ] || {
      printf "%s\n" "Remnawave Xray log is not available yet: $1" >&2
      exit 1
    }
    cat "$1"
  ' sh /var/log/supervisor/xray.out.log
}
require_debian() { [ -r /etc/debian_version ] && grep -q '^13' /etc/debian_version || { printf '%s\n' 'Debian 13 is required.' >&2; exit 1; }; command -v docker >/dev/null; command -v curl >/dev/null; command -v flock >/dev/null; command -v split >/dev/null; command -v stat >/dev/null; command -v tail >/dev/null; command -v truncate >/dev/null; command -v wc >/dev/null; systemctl is-active --quiet cron || { printf '%s\n' 'The cron service must be active.' >&2; exit 1; }; }
install_reporter() {
  install -d -m 700 /var/lib/tx-carpool-qps-detector
  install -m 600 /dev/null "$config_path"
  printf 'API_URL=%q\nNODE_TOKEN=%q\nINTERVAL_MINUTES=%q\n' "$api_url" "$node_token" "$interval" > "$config_path"
  cat > "$reporter_path" <<'REPORTER'
#!/usr/bin/env bash
set -euo pipefail
config_path=/etc/tx-carpool-qps-detector.conf
state_path=/var/lib/tx-carpool-qps-detector/report-state-v2
lock_path=/var/lib/tx-carpool-qps-detector/lock
log_path=/var/log/supervisor/xray.out.log
umask 077
exec 9>"$lock_path"
flock -n 9 || { [ "${REPORTER_REQUIRE_RUN:-0}" = 1 ] && exit 75; exit 0; }
source "$config_path"
started_at=$(date +%s)
last=0
offset=0
inode=
if [ -r "$state_path" ]; then read -r last offset inode < "$state_path" || true; fi
[[ "$last" =~ ^[0-9]+$ ]] || last=0
[[ "$offset" =~ ^[0-9]+$ ]] || offset=0
[ "$started_at" -ge "$((last + INTERVAL_MINUTES * 60))" ] || exit 0
payload=$(mktemp)
batch_dir=$(mktemp -d)
state_tmp="${state_path}.tmp"
trap 'rm -rf "$batch_dir"; rm -f "$payload" "$state_tmp"' EXIT
max_report_bytes=$((15 * 1024 * 1024))
metadata() {
  docker exec remnanode sh -c '
    if [ -e "$1" ]; then
      stat -c "%i %s" "$1"
    else
      printf "%s\n" "missing 0"
    fi
  ' sh "$log_path"
}
capture() {
  docker exec remnanode sh -c '
    [ ! -r "$1" ] || tail -c "$2" "$1"
  ' sh "$log_path" "+$((offset + 1))" > "$payload"
}
upload() {
  curl --fail --silent --show-error --connect-timeout 10 --max-time 120 --retry 2 --retry-all-errors --config - --data-binary "@$1" "$API_URL/api/v1/agents/qps-reports" > /dev/null <<CURL
header = "Authorization: Bearer $NODE_TOKEN"
request = "POST"
CURL
}
current_metadata=$(metadata)
read -r current_inode current_size <<< "$current_metadata"
[[ -n "$current_inode" && "$current_size" =~ ^[0-9]+$ ]] || {
  printf '%s\n' 'Could not read Remnawave Xray log metadata.' >&2
  exit 1
}
if [ -z "$inode" ]; then offset=$current_size; inode=$current_inode; fi
if [ "$inode" != "$current_inode" ] || [ "$offset" -gt "$current_size" ]; then offset=0; inode=$current_inode; fi
capture
after_metadata=$(metadata)
read -r after_inode after_size <<< "$after_metadata"
[[ -n "$after_inode" && "$after_size" =~ ^[0-9]+$ ]] || {
  printf '%s\n' 'Could not re-read Remnawave Xray log metadata.' >&2
  exit 1
}
if [ "$inode" != "$after_inode" ]; then offset=0; inode=$after_inode; capture; fi
captured_bytes=$(stat -c '%s' "$payload")
if [ "$captured_bytes" -gt 0 ] && [ "$(tail -c 1 "$payload" | wc -l)" -eq 0 ]; then
  partial_bytes=$(tail -n 1 "$payload" | wc -c)
  captured_bytes=$((captured_bytes - partial_bytes))
  truncate -s "$captured_bytes" "$payload"
fi
next_offset=$((offset + captured_bytes))
[ -s "$payload" ] || printf '\n' > "$payload"
split -C "$max_report_bytes" --numeric-suffixes=0 --suffix-length=4 "$payload" "$batch_dir/part-"
for part in "$batch_dir"/part-*; do upload "$part"; done
printf '%s %s %s\n' "$started_at" "$next_offset" "$inode" > "$state_tmp"
mv "$state_tmp" "$state_path"
REPORTER
  chmod 700 "$reporter_path"
  printf '* * * * * root %s\n' "$reporter_path" > "$cron_path"
  chmod 600 "$cron_path"
  if ! REPORTER_REQUIRE_RUN=1 "$reporter_path"; then
    printf '%s\n' 'Reporter installed, but the initial report was not accepted. Cron will retry.' >&2
    return 1
  fi
}
main() {
  require_root
  if [ "${1:-install}" = "uninstall" ]; then remove; return; fi
  if [ "${1:-install}" = "test" ]; then print_logs; return; fi
  require_debian
  read -r -p 'API domain (https://panel.example): ' api_url
  read -r -s -p 'Node token: ' node_token; printf '\n'
  read -r -p 'Report interval in minutes [30]: ' interval
  interval=${interval:-30}
  [[ "$api_url" =~ ^https?://[^[:space:]]+$ ]] && [[ "$node_token" =~ ^[A-Za-z0-9_-]{20,}$ ]] && [[ "$interval" =~ ^[1-9][0-9]{0,3}$ ]] || { printf '%s\n' 'Invalid configuration.' >&2; exit 1; }
  api_url=${api_url%/}
  install_reporter
  printf '%s\n' 'QPS detector reporter installed and the initial report was accepted.'
}
main "$@"
