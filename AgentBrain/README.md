# agent-brain-mcp

Stdio MCP server. Agents call tools that hit **agent-brain API** over HTTP.

## Setup

```bash
cp .env.example .env
npm install
npm run build
```

Cursor MCP example (API published on Docker host port 5090):

```json
{
  "agent-brain": {
    "command": "node",
    "args": ["C:/Users/armin/GitHub/agent-brain/mcp/dist/index.js"],
    "env": {
      "AGENT_BRAIN_API_URL": "http://127.0.0.1:5090",
      "AGENT_BRAIN_USERNAME": "armin",
      "AGENT_BRAIN_PASSWORD": "dopadopa123"
    }
  }
}
```

## Tools

- `health_check` — API health (SQLite + Qdrant)
- `list_sessions` — list Sessions (`q?` name search); JWT via env login

Knowledge store/search and session ingest use the HTTP API or `dispatcher-win`, not MCP tools.
