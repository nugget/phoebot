-- Verify phoebot:mapowner on pg

BEGIN;

    SELECT owner FROM map LIMIT 1;
    SELECT owner FROM poi LIMIT 1;

ROLLBACK;
