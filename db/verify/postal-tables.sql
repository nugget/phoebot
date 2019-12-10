-- Verify phoebot:postal-tables on pg

BEGIN;

    SELECT 1 FROM postal_scan LIMIT 1;

ROLLBACK;
