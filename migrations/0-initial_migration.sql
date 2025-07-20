BEGIN;

CREATE TABLE IF NOT EXISTS images(
    id TEXT PRIMARY KEY,
    image_data BYTEA NOT NULL,
    date TIMESTAMP,
    mimetype TEXT
);

COMMIT;
