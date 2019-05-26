-- Verify phoebot:function/onupdate on pg

BEGIN;

    SELECT has_function_privilege('onupdate_changed()', 'execute');

ROLLBACK;
