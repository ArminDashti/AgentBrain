# agent-brain

Centralized store for agent experience: general knowledge, expertise, resolved issues, task guides, and **Sessions** (Cursor agent chats from `state.vscdb`).

Accessible locally at **http://127.0.0.1:5091** (WebUI) and **http://127.0.0.1:5090** (API).

## Layout

| Folder | Role |
|--------|------|
| `api` | Gin API + SQLite + Qdrant |
| `webui` | Vue + Tailwind — read-only Sessions browse + stats (Dark+) |
| `mcp` | MCP stdio server (knowledge store/search via API) |
| `mcp-prompt` | MCP stdio server — tool `send-human-prompt` → Prompts API |
| `rule-skills` | Cursor plugin: rules + skills for store/search/ingest |
| `dispatcher-win` | Windows CLI: extract `state.vscdb` → API ingest (`ingest` / `sync`) |
| `scripts` | Windows Docker install (host ports 5090 / 5091) |

## Two surfaces

| Surface | Storage | Purpose |
|---------|---------|---------|
| **Knowledge** | SQLite `knowledge` + optional Qdrant | Durable knowledge kinds |
| **Sessions** | SQLite `sessions` + `turns` | Cursor session context %, tokens, thinking, turns (filled by Windows dispatcher) |

WebUI is **read-only** (session list/stats). Knowledge writes go through MCP/API. Session rows are filled by `dispatcher-win` from `%APPDATA%\Cursor\User\globalStorage\state.vscdb` (`sync` = all composers).

Knowledge kinds: `general_knowledge`, `expertise`, `resolved_issues`, `task_guide`.

## Quick start (Windows Docker)

```powershell
cd scripts
.\install-win-local-docker.ps1
```

Then open http://127.0.0.1:5091

Default login: `armin` / `dopadopa123`

- WebUI: `http://127.0.0.1:5091` (proxies `/api` to the API container)
- API: `http://127.0.0.1:5090`
- Qdrant: internal to compose (also used by API)

Fresh SQLite on first install. Use `.\reinstall-win-local-docker.ps1` to wipe volumes.

## Local API (without full stack)

```bash
cd api && cp .env.example .env && docker compose up -d
# Qdrant only; then:
go run ./cmd/server
```

## Storage choice (knowledge)

Agents pick `store_target` per write: `sqlite` (structured/filterable), `qdrant` (semantic), or `both`.

Session ingest always writes to SQLite (no Qdrant).

## Fill Sessions (Windows)

```powershell
cd dispatcher-win
go build -o agent-brain-dispatcher.exe ./cmd/agent-brain-dispatcher
.\agent-brain-dispatcher.exe sync
```

Then open WebUI → **Sessions** (`http://127.0.0.1:5091/sessions`).
