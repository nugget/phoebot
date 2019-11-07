-- Revert phoebot:containers from pg

BEGIN;

    DROP TABLE container;
    ALTER TABLE player DROP COLUMN verifyCode;
    ALTER TABLE channel DROP COLUMN playerID;

COMMIT;
