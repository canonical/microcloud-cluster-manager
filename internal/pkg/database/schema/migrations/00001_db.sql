-- +goose Up

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = CURRENT_TIMESTAMP;
   RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS remote_clusters (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL CHECK (status IN ('ACTIVE')),
    cluster_certificate TEXT NOT NULL UNIQUE,
    joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_remote_clusters_name ON remote_clusters (name);
CREATE INDEX IF NOT EXISTS idx_remote_clusters_certificate ON remote_clusters (cluster_certificate);
DROP TRIGGER IF EXISTS remote_clusters_updated_at_trigger ON remote_clusters;
CREATE TRIGGER remote_clusters_updated_at_trigger
    BEFORE UPDATE ON remote_clusters
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();


CREATE TABLE IF NOT EXISTS remote_cluster_tokens (
    id SERIAL PRIMARY KEY,
    encoded_token TEXT NOT NULL,
    expiry TIMESTAMPTZ NOT NULL DEFAULT '3000-01-01T00:00:00Z',
    cluster_name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_remote_cluster_tokens_name ON remote_cluster_tokens (cluster_name);


CREATE TABLE IF NOT EXISTS remote_cluster_details (
    id SERIAL PRIMARY KEY,
    remote_cluster_id INTEGER NOT NULL UNIQUE REFERENCES remote_clusters(id) ON DELETE CASCADE,
    cpu_total_count BIGINT NOT NULL DEFAULT 0,
    cpu_load_1 TEXT NOT NULL DEFAULT 0,
    cpu_load_5 TEXT NOT NULL DEFAULT 0,
    cpu_load_15 TEXT NOT NULL DEFAULT 0,
    memory_total_amount BIGINT NOT NULL DEFAULT 0,
    memory_usage BIGINT NOT NULL DEFAULT 0,
    disk_total_size BIGINT NOT NULL DEFAULT 0,
    disk_usage BIGINT NOT NULL DEFAULT 0,
    instance_count INTEGER NOT NULL DEFAULT 0,
    instance_statuses JSONB NOT NULL DEFAULT '[]'::JSONB,
    member_count INTEGER NOT NULL DEFAULT 0,
    member_statuses JSONB NOT NULL DEFAULT '[]'::JSONB,
    ui_url TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_remote_cluster_details_cluster_id ON remote_cluster_details (remote_cluster_id);
DROP TRIGGER IF EXISTS remote_cluster_details_updated_at_trigger ON remote_cluster_details;
CREATE TRIGGER remote_cluster_details_updated_at_trigger
    BEFORE UPDATE ON remote_cluster_details
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TABLE remote_cluster_config (
    id SERIAL           PRIMARY KEY,
    remote_cluster_id   INTEGER NOT NULL REFERENCES remote_clusters(id) ON DELETE CASCADE,
    key                 TEXT NOT NULL,
    value               TEXT NOT NULL,
    UNIQUE (remote_cluster_id, key)
);
