CREATE TABLE IF NOT EXISTS prompts (
    id TEXT PRIMARY KEY,
    prompt TEXT NOT NULL DEFAULT '',
    technically_revised TEXT NOT NULL DEFAULT '',
    english_revised TEXT NOT NULL DEFAULT '',
    harness TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_prompts_updated ON prompts (updated_at DESC);
