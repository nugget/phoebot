-- Revert phoebot:discordlogs from pg

BEGIN;

    DROP TABLE discordlog;
    DROP TABLE channel;

COMMIT;
