CREATE TABLE IF NOT EXISTS model_pricing (
    id TEXT PRIMARY KEY,
    model_id TEXT NOT NULL,
    provider TEXT NOT NULL,
    input_cost_per_1m REAL NOT NULL DEFAULT 0,
    output_cost_per_1m REAL NOT NULL DEFAULT 0,
    image_cost_per_unit REAL NOT NULL DEFAULT 0,
    currency TEXT NOT NULL DEFAULT 'USD',
    effective_date TEXT NOT NULL DEFAULT (date('now')),
    source TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(model_id, provider)
);

CREATE INDEX IF NOT EXISTS idx_model_pricing_model ON model_pricing(model_id);
CREATE INDEX IF NOT EXISTS idx_model_pricing_provider ON model_pricing(provider);

CREATE TABLE IF NOT EXISTS tags (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    color TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS model_tags (
    model_id TEXT NOT NULL,
    tag_id TEXT NOT NULL,
    PRIMARY KEY (model_id, tag_id),
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS combos (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    strategy TEXT NOT NULL DEFAULT 'fallback',
    sticky_limit INTEGER NOT NULL DEFAULT 0,
    enabled INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS combo_models (
    id TEXT PRIMARY KEY,
    combo_id TEXT NOT NULL,
    model TEXT NOT NULL,
    priority INTEGER NOT NULL DEFAULT 0,
    weight INTEGER NOT NULL DEFAULT 1,
    enabled INTEGER NOT NULL DEFAULT 1,
    FOREIGN KEY (combo_id) REFERENCES combos(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_combo_models_combo ON combo_models(combo_id);
