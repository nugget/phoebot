-- Revert phoebot:containers from pg

BEGIN;

    DROP TABLE container;
    ALTER TABLE player DROP COLUMN verifyCode;
    ALTER TABLE player DROP COLUMN verified;
    ALTER TABLE channel DROP COLUMN playerID;

COMMIT;
