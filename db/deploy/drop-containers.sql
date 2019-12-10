-- Deploy phoebot:drop-containers to pg

BEGIN;

    DROP TABLE container;

COMMIT;
