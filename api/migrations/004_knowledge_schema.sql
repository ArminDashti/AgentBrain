CREATE TABLE IF NOT EXISTS knowledge__v2 (
    id TEXT PRIMARY KEY,
    kind TEXT NOT NULL CHECK (kind IN ('general_knowledge', 'expertise', 'resolved_issues', 'task_guide')),
    title TEXT NOT NULL,
    body TEXT NOT NULL DEFAULT '',
    tags TEXT NOT NULL DEFAULT '[]',
    metadata TEXT NOT NULL DEFAULT '{}',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

INSERT OR IGNORE INTO knowledge__v2 (id, kind, title, body, tags, metadata, created_at, updated_at)
SELECT id, kind, title, body, tags, metadata, created_at, updated_at FROM knowledge;

DROP TABLE IF EXISTS knowledge;
ALTER TABLE knowledge__v2 RENAME TO knowledge;

CREATE INDEX IF NOT EXISTS idx_knowledge_kind ON knowledge (kind);
CREATE INDEX IF NOT EXISTS idx_knowledge_updated ON knowledge (updated_at DESC);
