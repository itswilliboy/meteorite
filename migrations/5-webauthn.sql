BEGIN;

CREATE TABLE IF NOT EXISTS webauthn_credentials (
    id BYTEA PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    name TEXT NOT NULL,
    data JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT current_timestamp
);

COMMIT;
