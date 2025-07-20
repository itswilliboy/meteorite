BEGIN;

CREATE TABLE IF NOT EXISTS users(
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE,
    created_at TIMESTAMP DEFAULT current_timestamp,
    enabled BOOLEAN DEFAULT false
);

CREATE TABLE IF NOT EXISTS tokens(
    user_id INT PRIMARY KEY REFERENCES users(id),
    token BYTEA NOT NULL
);

ALTER TABLE images ADD user_id INTEGER;
ALTER TABLE images
    ADD CONSTRAINT fk_images_users
    FOREIGN KEY (user_id)
    REFERENCES users(id);

COMMIT;
