-- +goose Up
CREATE TABLE IF NOT EXISTS identities (
    id SERIAL PRIMARY KEY,
    subject VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    display_name VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_seen TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_identities_email ON identities (email);

CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY,
    identity_id INTEGER NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    id_token_encrypted TEXT NOT NULL,
    access_token_encrypted TEXT NOT NULL,
    refresh_token_encrypted TEXT NOT NULL,
    FOREIGN KEY (identity_id) REFERENCES identities (id) ON DELETE CASCADE
);


-- +goose Down
DROP TABLE IF EXISTS sessions;

DROP INDEX IF EXISTS idx_identities_email;
DROP TABLE IF EXISTS identities;
