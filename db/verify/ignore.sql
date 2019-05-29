-- Verify phoebot:ignore on pg

BEGIN;

    SELECT 1 FROM ignore LIMIT 1;

ROLLBACK;
