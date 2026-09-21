-- +goose Up
ALTER TABLE remote_clusters
    ADD COLUMN cluster_uuid TEXT NOT NULL DEFAULT '';


-- +goose Down
ALTER TABLE remote_clusters
    DROP COLUMN cluster_uuid;
