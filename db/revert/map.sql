-- Revert phoebot:map from pg

BEGIN;

    DROP TABLE poi;
    DROP TABLE map;

COMMIT;
