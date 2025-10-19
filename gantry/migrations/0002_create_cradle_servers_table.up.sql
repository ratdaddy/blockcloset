CREATE TABLE IF NOT EXISTS cradle_servers (
    id TEXT PRIMARY KEY,
    address TEXT NOT NULL UNIQUE,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);
