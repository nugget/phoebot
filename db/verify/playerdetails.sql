-- Verify phoebot:playerdetails on pg

BEGIN;

    SELECT username, locale FROM player LIMIT 1;

ROLLBACK;
