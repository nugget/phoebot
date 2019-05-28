-- Revert phoebot:mojangreleases from pg

BEGIN;

    DROP TABLE mojangnews;

COMMIT;
