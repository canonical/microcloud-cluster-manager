-- +goose Up
CREATE TABLE IF NOT EXISTS user_access_tokens (
    id SERIAL PRIMARY KEY,
    access_token TEXT NOT NULL DEFAULT ''
);


-- +goose Down
DROP TABLE user_access_tokens;
