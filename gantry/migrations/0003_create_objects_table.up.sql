CREATE TABLE IF NOT EXISTS objects (
    object_id TEXT PRIMARY KEY,
    bucket_id TEXT NOT NULL,
    key TEXT NOT NULL,
    state TEXT NOT NULL CHECK (state IN ('PENDING','COMMITTED','FAILED','REPLACED')),
    size_expected INTEGER NOT NULL,
    size_actual INTEGER,
    last_modified INTEGER,
    cradle_server_id TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    FOREIGN KEY (bucket_id) REFERENCES buckets(id) ON DELETE RESTRICT,
    FOREIGN KEY (cradle_server_id) REFERENCES cradle_servers(id) ON DELETE RESTRICT
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_objects_committed_unique
    ON objects(bucket_id, key) WHERE state = 'COMMITTED';

CREATE INDEX IF NOT EXISTS idx_objects_bucket_key
    ON objects(bucket_id, key);
