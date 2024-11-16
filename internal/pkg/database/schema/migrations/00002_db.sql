-- TODO: this is just an example!
-- +goose Up
ALTER TABLE remote_clusters
ADD COLUMN age INTEGER DEFAULT 0;

-- +goose Down
ALTER TABLE remote_clusters
DROP COLUMN age;