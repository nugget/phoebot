-- Revert phoebot:postal-tables from pg

BEGIN;

    DROP TABLE postal_scan;

COMMIT;
