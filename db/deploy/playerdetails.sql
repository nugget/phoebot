-- Deploy phoebot:playerdetails to pg

BEGIN;

    ALTER TABLE player ADD COLUMN username varchar;
    ALTER TABLE player ADD COLUMN locale varchar;

COMMIT;
