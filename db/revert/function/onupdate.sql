-- Revert phoebot:function/onupdate from pg

BEGIN;

    DROP FUNCTION onupdate_changed();

COMMIT;
