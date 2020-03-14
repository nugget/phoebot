-- Revert phoebot:verify-nags from pg

BEGIN;

    DROP TABLE IF EXISTS verify;

    ALTER TABLE player ADD COLUMN timezone varchar;
    ALTER TABLE player ADD COLUMN locale varchar;
    ALTER TABLE player ADD COLUMN ignored boolean;
    ALTER TABLE player ADD COLUMN verifycode varchar;

COMMIT;
