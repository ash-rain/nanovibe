package db

const schema = `
CREATE TABLE IF NOT EXISTS setup_state (
    id             INTEGER PRIMARY KEY,
    current_step   TEXT    NOT NULL DEFAULT 'welcome',
    completed_steps TEXT   NOT NULL DEFAULT '[]',
    updated_at     INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS projects (
    id               TEXT    PRIMARY KEY,
    name             TEXT    NOT NULL,
    path             TEXT    NOT NULL UNIQUE,
    language         TEXT,
    github_url       TEXT,
    git_remote       TEXT,
    default_provider TEXT,
    created_at       INTEGER NOT NULL,
    last_opened_at   INTEGER
);

CREATE TABLE IF NOT EXISTS settings (
    key        TEXT    PRIMARY KEY,
    value      TEXT    NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS agent_messages (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    role       TEXT    NOT NULL,
    content    TEXT    NOT NULL,
    created_at INTEGER NOT NULL,
    project_id TEXT
);

CREATE TABLE IF NOT EXISTS github_auth (
    id           INTEGER PRIMARY KEY,
    login        TEXT,
    avatar_url   TEXT,
    public_repos INTEGER,
    connected_at INTEGER
);
`

// seedSQL is run after schema creation to ensure singleton rows exist.
const seedSQL = `
INSERT OR IGNORE INTO setup_state (id, current_step, completed_steps, updated_at)
VALUES (1, 'welcome', '[]', strftime('%s','now'));

INSERT OR IGNORE INTO github_auth (id) VALUES (1);
`
