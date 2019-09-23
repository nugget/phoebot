-- Verify phoebot:poi-extras on pg

BEGIN;

    SELECT dimension FROM poi LIMIT 1;

ROLLBACK;
