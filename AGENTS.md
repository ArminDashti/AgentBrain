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

### 2026-11-14 — Cloud Agent dev stack (SQLite, ports from .env)

**Goal.** On a fresh Cloud Agent VM, run the API and Vite WebUI so an agent can log in and store a prompt or knowledge row in SQLite.

**Steps.**
- Install Go **1.26.10** to `/usr/local/go` and symlink `go` and `gofmt` into `/usr/local/bin` (`scripts/cloud-agent-install.sh`). Use `GOTOOLCHAIN=local` and `CGO_ENABLED=0`. `modernc.org/sqlite` builds without cgo. Node 22 is already on the login-shell `PATH` (`bash -lc`).
- `go mod download` in `api` and `dispatcher-win`. `npm ci` in `webui` and `AgentBrain`, then `npm run build` in `AgentBrain` (`dist/` is gitignored).
- If `.env` is missing, copy `.env.example` in `api`, `webui`, and `AgentBrain`. Do not overwrite an edited `.env`.
- From repo root: `./scripts/run-local-dev.sh`. It reads `ADDR` and `VITE_DEV_PORT` from those env files. The examples bind API **http://127.0.0.1:8220** and WebUI **http://127.0.0.1:5220**.
- Wait until `http://127.0.0.1:8220/health` is HTTP 200 and `"ok": true`, and the WebUI returns a non-error HTTP status. The Vite server proxies `/api` and `/health` to the API.
- Login credentials are the API seed in `api/internal/db` (`DefaultUsername`, `DefaultPassword`). Do not copy the password into this file.

**Gotchas.**
- Docker host ports **5090** (API) and **5091** (WebUI) come from `scripts/docker-compose.yml` only. They are not the Vite/Gin dev ports.
- `config.LoadDotEnv` does not override variables already set in the process. `set -a; . ./.env` does overwrite earlier shell exports.
- `go test ./...` in `api` fails: `internal/http` tests import removed packages `internal/embed` and `internal/qdrant`. `go build -o /tmp/agent-brain-api ./cmd/server` and `go test ./internal/timefmt` pass. `dispatcher-win` has no tests. `npm run build` in `webui` passes.

**What didn't work.**
- System Go 1.22.2 (`/usr/bin/go`) cannot compile `api` (`go 1.25.0`) or `dispatcher-win` (`go 1.26.6`). `./scripts/run-local-dev.sh` sets `GOTOOLCHAIN=local`, so it will not fetch a newer toolchain. Run `./scripts/cloud-agent-install.sh` first; the dev script now exits immediately when `go version` is older than 1.25.
- `./scripts/run-local-dev.sh` before 2026-11-14 exited immediately: `cp mcp-prompt/.env.example` failed because that directory is `AgentBrain`.
- Checking health on **5090** / **5091** after copying `.env.example` never becomes ready: those files set `ADDR=:8220` and `VITE_DEV_PORT=5220`.
