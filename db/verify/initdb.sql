-- Verify phoebot:initdb on pg

BEGIN;

    SELECT 1 FROM config LIMIT 1;
    SELECT 1 FROM player LIMIT 1;
    SELECT 1 FROM acl LIMIT 1;
    SELECT 1 FROM subscription LIMIT 1;
    SELECT 1 FROM product LIMIT 1;

ROLLBACK;
