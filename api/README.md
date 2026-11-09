# agent-brain-api

Gin + SQLite knowledge API.

## Run

```bash
cp .env.example .env
go run ./cmd/server
```

Env: `SQLITE_PATH` (default `./data/agent_brain.db`), `JWT_SECRET`.

## Knowledge

`store_target`: `sqlite` (only).

Search `mode`: `structured` (SQLite LIKE on title/body).

Kinds: `general_knowledge`, `expertise`, `resolved_issues`, `task_guide`.
