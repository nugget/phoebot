-- Revert phoebot:mailboxes from pg

BEGIN;

    DROP TABLE mailbox;

COMMIT;
