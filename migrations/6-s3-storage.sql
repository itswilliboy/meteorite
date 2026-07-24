BEGIN;

ALTER TABLE media ADD COLUMN size BIGINT;
ALTER TABLE media ADD COLUMN has_cover BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE media ALTER COLUMN data DROP NOT NULL;

UPDATE media SET size = octet_length(data), has_cover = (cover_art IS NOT NULL);

COMMIT;
