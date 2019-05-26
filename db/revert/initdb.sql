-- Revert phoebot:initdb from pg

BEGIN;

    DROP TABLE config;
    DROP TABLE player;
    DROP TABLE acl;
    DROP TABLE subscription;
    DROP TABLE product;

COMMIT;
