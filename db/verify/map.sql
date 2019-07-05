-- Verify phoebot:map on pg

BEGIN;

    SELECT 1 FROM map LIMIT 1;
    SELECT 1 FROM poi LIMIT 1;

ROLLBACK;
