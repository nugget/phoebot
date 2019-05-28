-- Verify phoebot:discordlogs on pg

BEGIN;

    SELECT 1 FROM channel LIMIT 1;

ROLLBACK;
