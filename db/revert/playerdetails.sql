-- Revert phoebot:playerdetails from pg

BEGIN;

    ALTER TABLE PLAYER DROP COLUMN locale;
    ALTER TABLE PLAYER DROP COLUMN username;

COMMIT;
