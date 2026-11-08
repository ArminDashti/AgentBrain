#!/usr/bin/env bash
# Local dev on Linux/macOS (Cloud Agent VM). Matches README ports 5090 (API) / 5091 (WebUI).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
TMUX_CONF="${TMUX_CONF:-/exec-daemon/tmux.portal.conf}"
tmux() { command tmux -f "$TMUX_CONF" "$@"; }

export GOTOOLCHAIN=local CGO_ENABLED=0
export PATH="/usr/local/go/bin:${PATH:-}"

QDRANT_CFG="${HOME}/.local/share/agent-brain/qdrant-config.yaml"
mkdir -p "${HOME}/.local/share/agent-brain/qdrant" "${HOME}/.local/share/agent-brain/sqlite"

if [[ ! -f "$QDRANT_CFG" ]]; then
  cat >"$QDRANT_CFG" <<EOF
storage:
  storage_path: ${HOME}/.local/share/agent-brain/qdrant
service:
  http_port: 6335
  grpc_port: 6336
EOF
fi

for d in api webui mcp-prompt; do
  [[ -f "$ROOT/$d/.env" ]] || cp "$ROOT/$d/.env.example" "$ROOT/$d/.env"
done

start_session() {
  local name=$1 dir=$2 cmd=$3
  if tmux has-session -t "=$name" 2>/dev/null; then
    return 0
  fi
  tmux new-session -d -s "$name" -c "$dir" -- bash -lc "$cmd"
}

if ! curl -sf http://127.0.0.1:6335/readyz >/dev/null; then
  start_session qdrant "$HOME" 'exec qdrant --config-path "'"$QDRANT_CFG"'"'
fi

if ! curl -sf http://127.0.0.1:5090/health >/dev/null 2>&1; then
  start_session api "$ROOT/api" '
    export SQLITE_PATH="'"${HOME}"'/.local/share/agent-brain/sqlite/agent_brain.db"
    set -a && [ -f .env ] && . ./.env && set +a
    exec go run ./cmd/server
  '
fi

if ! curl -sf -o /dev/null http://127.0.0.1:5091/ 2>/dev/null; then
  start_session webui "$ROOT/webui" '
    set -a && [ -f .env ] && . ./.env && set +a
    exec npm run dev -- --host 0.0.0.0 --port "${VITE_DEV_PORT:-5091}"
  '
fi

echo "Waiting for services..."
for _ in $(seq 1 90); do
  curl -sf http://127.0.0.1:6335/readyz >/dev/null || { sleep 1; continue; }
  curl -sf http://127.0.0.1:5090/health | grep -q '"ok":true' || { sleep 1; continue; }
  code=$(curl -sf -o /dev/null -w "%{http_code}" http://127.0.0.1:5091/ || echo 000)
  [[ "$code" != "000" ]] || { sleep 1; continue; }
  echo "Ready:"
  echo "  WebUI  http://127.0.0.1:5091"
  echo "  API    http://127.0.0.1:5090/health"
  echo "  Login  armin / (see README)"
  exit 0
done
echo "Timed out; check: tmux -f $TMUX_CONF attach -t api" >&2
exit 1
