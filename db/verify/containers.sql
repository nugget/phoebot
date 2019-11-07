-- Verify phoebot:containers on pg

BEGIN;

    SELECT 1 FROM container LIMIT 1;
    SELECT verifycode FROM player LIMIT 1;
    SELECT playerID FROM channel LIMIT 1;

ROLLBACK;
