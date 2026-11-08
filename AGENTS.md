# Self-learning (for AI coding agents)

This file makes any coding agent **self-improving**: recognize a hard-won
"golden path" during a task and persist it so the next session starts already
knowing it, instead of rediscovering how to reach the DB, where the creds live,
how to deploy, or how to verify a change live.

It works with any agent that reads a standing instructions file (Codex, Zed,
Aider, Gemini CLI, …). Richer, tool-native installs exist too — a Claude Code
**skill** (`skills/self-learning/SKILL.md`) and a **Cursor rule**
(`.cursor/rules/self-learning.mdc`); see the README. This file is the portable,
lowest-common-denominator version.

## The loop

**1. Recognize the moment.** Any one of these is a cue:
- a task only worked after several attempts, wrong turns, or a correction;
- you discovered project facts you didn't know up front — where creds/env vars
  live, a non-obvious command, a required sequence, a gotcha;
- an operational workflow likely to recur (reach the dev/prod DB, deploy, run
  migrations, seed data, verify live, tail the right logs);
- the user says "remember this" / "don't make me re-explain this next time".

Act on the cue immediately — **don't ask permission first**. Capture it, then
tell the user what you saved and where. They can always edit or delete it.

**2. Capture it where your tool auto-loads knowledge next session:**
- Claude Code / any Agent Skills client → a new `skills/<name>/SKILL.md`
- Cursor → a new `.cursor/rules/learned/<name>.mdc`
- Otherwise → append a dated entry under [Learned](#learned) below, or to your
  project's notes/memory file.

Capture the **procedure** (commands, paths, the required order, gotchas) — not a
one-off answer — and the **failures** too: the approaches you ruled out and why,
so next time skips the dead-ends.

**3. Reuse.** Next session the persisted entry loads automatically (by skill/rule
description, or because this file is always read) and you start from the golden
path.

## Promotion rule

A saved entry is authoritative — future sessions trust it without re-deriving it.
Only promote a session to a durable entry when **all three** hold:

1. **A passing check** — the path was actually verified (a test passed, the
   command exited clean, the repro reproduced, the build went green). Record it.
   "Seemed to work" doesn't count.
2. **A named failure pattern** — you can name the failure it avoids or diagnoses,
   not a vague "sometimes it breaks".
3. **At least one ruled-out dead-end** — a concrete approach you tried and
   eliminated, with the reason.

If any is missing, it isn't durable yet — leave a tentative note (marked
unverified) or skip it. This keeps confident guesses out.

## Rules

- **Never write secret values** — no tokens, passwords, connection strings, or
  API keys. Record only *where* a secret lives (env var name, config/selector,
  secret manager). Reproducing a secret into a shared file leaks it.
- **A one-line fact or correction** → put it in lightweight notes/memory, not a
  whole rule or skill.
- **A genuine one-off** unlikely to recur → skip it.
- **Capture procedures, not answers** — teach how to approach the class of
  problem, so it generalizes next time.

## Learned

<!-- When no richer mechanism is available, append dated golden-path entries here.
     Format: ### YYYY-MM-DD — <title>  /  **Goal**, **Steps**, **Gotchas**,
     **What didn't work**. Keep secrets out — point to where they live. -->

### 2026-11-06 — Cloud Agent dev stack

**Goal.** On a fresh Cloud Agent VM, run Qdrant, the API, and the Vite WebUI so an agent can log in, write a prompt, and store knowledge in SQLite and Qdrant.

**Steps.**
- The default image Go is too old. Install Go **1.26.10** (satisfies `dispatcher-win/go.mod` `go 1.26.6` and `api/go.mod` `go 1.25.0`) to `/usr/local/go` and symlink `go` and `gofmt` into `/usr/local/bin`. Use `GOTOOLCHAIN=local` and `CGO_ENABLED=0`. `modernc.org/sqlite` builds without cgo.
- Install Qdrant **v1.13.2** (the compose pin) from `qdrant-x86_64-unknown-linux-musl.tar.gz` to `/usr/local/bin/qdrant`. Start with a small config at `$HOME/.local/share/agent-brain/qdrant-config.yaml` (`storage.storage_path` → `$HOME/.local/share/agent-brain/qdrant`, `service.http_port` **6335**, `grpc_port` **6336**): `qdrant --config-path "$HOME/.local/share/agent-brain/qdrant-config.yaml"`. That host port is what `api/.env.example` already sets (`QDRANT_URL=http://127.0.0.1:6335`); compose publishes `6335:6333`.
- If `.env` is missing, copy `.env.example` in `api`, `webui`, `mcp`, and `mcp-prompt`. When `mcp` or `mcp-prompt` still has `AGENT_BRAIN_API_URL=http://127.0.0.1:5090`, change only that default to `http://127.0.0.1:8220`. Do not overwrite an edited `.env`.
- `go mod download` in `api` and `dispatcher-win`. `npm ci` in `webui`, `mcp`, and `mcp-prompt`, then `npm run build` in the two MCP packages. Node 22 is already on the login-shell `PATH`; do not install another Node.
- Start order, each in its own tmux session (`qdrant`, `api`, `webui`), skipping a session when its port already answers:
  1. Qdrant, then `curl -sf http://127.0.0.1:6335/readyz` (`all shards are ready`).
  2. From `api/`: `go run ./cmd/server` with `SQLITE_PATH=$HOME/.local/share/agent-brain/sqlite/agent_brain.db` so the database stays outside the worktree. Wait until `http://127.0.0.1:8220/health` is HTTP 200 and both `sqlite` and `qdrant` are true.
  3. From `webui/`: `npm run dev -- --host 0.0.0.0 --port 5220`. Open `http://127.0.0.1:5220` (proxies `/api` to `:8220`).
- Login credentials are the API seed in `api/internal/db` (`DefaultUsername`, `DefaultPassword`). Do not copy the password into this file.
- Checks that passed on 2026-11-06: `go test ./...` in `api` (timefmt ok), `npm run build` in `webui`, login plus `POST /api/v1/knowledge` with `store_target` `both` (`in_sqlite` and `in_qdrant` true), `POST /api/v1/prompts`, and the WebUI flow Stats → Prompts (create + save) → Sessions.

**Gotchas.**
- Startup only logs a warning when Qdrant is down, but `GET /health` returns **503** until Qdrant is ready. Wait on `/health` 200, not on the process listening.
- Root README ports **5090** / **5091** are the Windows Docker compose stack (`scripts/docker-compose.yml`). Local dev is API **8220** and WebUI **5220**.
- `npm ci` is idempotent and exits 0 on a second run; it still reinstalls `node_modules`.

**What didn't work.**
- `/usr/bin/go` **1.22.2** cannot compile these modules. The `go 1.26.6` line fails the toolchain check before any package builds.
- Docker Compose for Qdrant (the README path) was not used. Docker is not on the default image, and the API only needs the HTTP port already in `api/.env.example`. The v1.13.2 binary on port 6335 is the verified path.
- Qdrant **1.13.2** rejects CLI `--storage-path`; use `storage.storage_path` in a config file (or env overrides) instead.
