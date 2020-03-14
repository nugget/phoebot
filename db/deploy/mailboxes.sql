-- Deploy phoebot:mailboxes to pg

BEGIN;

    CREATE TABLE mailbox (
        mailboxID uuid NOT NULL DEFAULT gen_random_uuid(),
        added timestamp(0) NOT NULL DEFAULT current_timestamp,
        changed timestamp(0) NOT NULL DEFAULT current_timestamp,
        deleted timestamp(0),
        lastscan timestamp(0) NOT NULL DEFAULT current_timestamp,
        enabled boolean NOT NULL DEFAULT TRUE,
        class varchar NOT NULL DEFAULT 'mailbox',
        owner varchar NOT NULL DEFAULT '',
        signtext varchar NOT NULL DEFAULT '',
        material varchar NOT NULL DEFAULT '',
        world varchar NOT NULL DEFAULT 'world',
        x int,
        y int,
        z int,
        flag boolean NOT NULL DEFAULT FALSE,
        PRIMARY KEY(mailboxID)
    );
    CREATE TRIGGER onupdate BEFORE UPDATE ON mailbox FOR EACH ROW EXECUTE PROCEDURE onupdate_changed();
    GRANT SELECT, INSERT, UPDATE ON mailbox TO phoebot;

    DROP TABLE IF EXISTS scanrange;

COMMIT;
