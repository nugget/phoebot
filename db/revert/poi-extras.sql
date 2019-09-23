-- Revert phoebot:poi-extras from pg

BEGIN;

    ALTER TABLE poi DROP COLUMN dimension;

COMMIT;
