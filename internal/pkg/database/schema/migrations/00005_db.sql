-- +goose Up
ALTER TABLE remote_cluster_details
RENAME COLUMN ui_url TO lxd_url;

-- +goose Down
ALTER TABLE remote_cluster_details
RENAME COLUMN lxd_url TO ui_url;
