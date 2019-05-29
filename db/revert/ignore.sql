-- Revert phoebot:ignore from pg

BEGIN;

    DROP TABLE ignore;

COMMIT;
