-- Revert phoebot:mapowner from pg

BEGIN;

    ALTER TABLE map DROP COLUMN owner;
    ALTER TABLE poi DROP COLUMN owner;

COMMIT;
