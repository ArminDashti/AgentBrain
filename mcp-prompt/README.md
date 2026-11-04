# agent-brain-prompt-mcp

Stdio MCP server with one tool: **`send-human-prompt`**. Agents submit human prompts to **agent-brain API** (`POST /api/v1/prompts`).

## Setup

```bash
cp .env.example .env
npm install
npm run build
```

Windows (registers `agent-brain-prompt` in Cursor `mcp.json`):

```powershell
cd scripts
.\install-win.ps1
```

Cursor MCP example:

```json
{
  "agent-brain-prompt": {
    "command": "node",
    "args": ["C:/Users/armin/GitHub/AgentBrain/mcp-prompt/dist/index.js"],
    "env": {
      "AGENT_BRAIN_API_URL": "http://127.0.0.1:5090",
      "AGENT_BRAIN_USERNAME": "armin",
      "AGENT_BRAIN_PASSWORD": "dopadopa123"
    }
  }
}
```

## Tool: `send-human-prompt`

| Argument | Required | Description |
|----------|----------|-------------|
| `prompt` | yes | Raw human prompt |
| `technically_revised` | no | Technical rewrite (default empty) |
| `english_revised` | no | English rewrite (default empty) |
| `harness` | no | Harness label |

Returns JSON `{ "id": "<uuid>" }`. View prompts in WebUI → **Prompts**.
