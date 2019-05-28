-- Revert phoebot:discordlogs from pg

BEGIN;

    DROP TABLE channel;

COMMIT;
