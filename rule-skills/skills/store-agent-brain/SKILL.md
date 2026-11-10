---
name: store-agent-brain
description: Store experience, resolved issues, task guides, or expertise into agent-brain via HTTP API.
---

# Store into agent-brain

Use agent-brain HTTP API (not MCP):

1. `POST /api/v1/auth/login` with username/password (env defaults `armin` / `dopadopa123`).
2. `POST /api/v1/knowledge` with Bearer token and the fields below.

## Fields

| Field | Values |
|-------|--------|
| `kind` | `general_knowledge` \| `expertise` \| `resolved_issues` \| `task_guide` |
| `title` | Short label |
| `body` | Full text |
| `tags` | Optional string array |
| `store_target` | `sqlite` \| `qdrant` \| `both` |

## Target choice

- Structured inventory / filters → `sqlite`
- “Find similar later” → `qdrant` or `both`
- Default for important fixes → `both`
