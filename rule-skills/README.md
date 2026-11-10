# agent-brain-rule-skills

Cursor plugin: rules + skills for the **agent-brain** project.

- Knowledge: HTTP `POST /api/v1/knowledge` / `POST /api/v1/knowledge/search`
- Sessions: Windows `dispatcher-win` `sync` / `ingest` fills from `state.vscdb`; MCP `list_sessions` to browse

Install as a local Cursor plugin pointing at this folder, or copy rules/skills into a workspace that should use agent-brain.

Does not auto-wire into armin-command-center.
