-- Verify phoebot:mailboxes on pg

BEGIN;

    SELECT 1 FROM mailbox LIMIT 1;

ROLLBACK;
