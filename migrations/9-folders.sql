BEGIN;

CREATE TABLE folders (
    id TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    parent_id TEXT REFERENCES folders(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT current_timestamp
);

CREATE INDEX folders_user_parent_idx ON folders (user_id, parent_id);

ALTER TABLE media ADD COLUMN folder_id TEXT REFERENCES folders(id) ON DELETE CASCADE;
CREATE INDEX media_user_folder_idx ON media (user_id, folder_id);

COMMIT;
