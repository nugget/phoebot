-- Verify phoebot:verify-nags on pg

BEGIN;

    SELECT 1 FROM verify LIMIT 1;

ROLLBACK;
