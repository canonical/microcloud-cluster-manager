-- +goose Up
ALTER TABLE remote_cluster_details
    ADD COLUMN tunnel_manager_member_url TEXT NOT NULL DEFAULT '';


-- +goose Down
ALTER TABLE remote_cluster_details
    DROP COLUMN tunnel_manager_member_url;
