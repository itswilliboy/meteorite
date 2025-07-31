ALTER TABLE users ADD admin BOOLEAN DEFAULT false;
ALTER TABLE users ADD password BYTEA NOT NULL;

ALTER TABLE images ADD views INTEGER NOT NULL; -- timestamp[]
