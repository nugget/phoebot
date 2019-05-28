-- Verify phoebot:mojangreleases on pg

BEGIN;

    SELECT 1 FROM mojangnews LIMIT 1;

ROLLBACK;
