---
name: search-agent-brain
description: Search agent-brain for prior resolved issues, task guides, and expertise via HTTP API.
---

# Search agent-brain

Use agent-brain HTTP API (not MCP): login, then `POST /api/v1/knowledge/search`.

## Mode

| Mode | Backend |
|------|---------|
| `structured` | SQLite LIKE on title and body (default) |

Optional `kind` filter: `general_knowledge`, `expertise`, `resolved_issues`, `task_guide`.

Also available: `GET /api/v1/knowledge` (list), `GET /api/v1/knowledge/{id}` (get).
