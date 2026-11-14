#!/usr/bin/env bash
# Local dev on Linux/macOS (Cloud Agent VM).
# Ports come from api/.env and webui/.env (examples: API :8220, WebUI :5220).
# Docker host ports 5090/5091 are only for scripts/docker-compose.yml.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
TMUX_CONF="${TMUX_CONF:-/exec-daemon/tmux.portal.conf}"
if [[ -f "$TMUX_CONF" ]]; then
  tmux() { command tmux -f "$TMUX_CONF" "$@"; }
else
  tmux() { command tmux "$@"; }
fi

export GOTOOLCHAIN=local CGO_ENABLED=0
export PATH="/usr/local/go/bin:${PATH:-}"

for d in api webui AgentBrain; do
  if [[ -f "$ROOT/$d/.env.example" && ! -f "$ROOT/$d/.env" ]]; then
    cp "$ROOT/$d/.env.example" "$ROOT/$d/.env"
  fi
done

set -a
# shellcheck disable=SC1091
[[ -f "$ROOT/api/.env" ]] && . "$ROOT/api/.env"
# shellcheck disable=SC1091
[[ -f "$ROOT/webui/.env" ]] && . "$ROOT/webui/.env"
set +a
API_PORT="${ADDR:-:8220}"
API_PORT="${API_PORT##*:}"
WEBUI_PORT="${VITE_DEV_PORT:-5220}"

start_session() {
  local name=$1 dir=$2 cmd=$3
  if tmux has-session -t "=$name" 2>/dev/null; then
    tmux kill-session -t "=$name"
  fi
  tmux new-session -d -s "$name" -c "$dir" -- bash -lc "$cmd"
}

if ! curl -sf "http://127.0.0.1:${API_PORT}/health" >/dev/null 2>&1; then
  start_session api "$ROOT/api" '
    export GOTOOLCHAIN=local CGO_ENABLED=0
    export PATH="/usr/local/go/bin:${PATH:-}"
    set -a && [ -f .env ] && . ./.env && set +a
    exec go run ./cmd/server
  '
fi

if ! curl -sf -o /dev/null "http://127.0.0.1:${WEBUI_PORT}/" 2>/dev/null; then
  start_session webui "$ROOT/webui" "
    set -a && [ -f .env ] && . ./.env && set +a
    exec npm run dev -- --host 0.0.0.0 --port \"\${VITE_DEV_PORT:-${WEBUI_PORT}}\"
  "
fi

echo "Waiting for services..."
for _ in $(seq 1 120); do
  curl -sf "http://127.0.0.1:${API_PORT}/health" | grep -q '"ok":true' || { sleep 1; continue; }
  code=$(curl -sf -o /dev/null -w "%{http_code}" "http://127.0.0.1:${WEBUI_PORT}/" || echo 000)
  [[ "$code" != "000" ]] || { sleep 1; continue; }
  echo "Ready:"
  echo "  WebUI  http://127.0.0.1:${WEBUI_PORT}"
  echo "  API    http://127.0.0.1:${API_PORT}/health"
  exit 0
done
echo "Timed out waiting for API :${API_PORT} and WebUI :${WEBUI_PORT}" >&2
if [[ -f "$TMUX_CONF" ]]; then
  echo "Check: tmux -f $TMUX_CONF attach -t api" >&2
else
  echo "Check: tmux attach -t api" >&2
fi
exit 1
