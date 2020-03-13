-- Deploy phoebot:postal-tables to pg

BEGIN;

    CREATE TABLE scanrange (
        scanrangeID uuid NOT NULL DEFAULT gen_random_uuid(),
        added timestamp(0) NOT NULL DEFAULT current_timestamp,
        changed timestamp(0) NOT NULL DEFAULT current_timestamp,
        deleted timestamp(0),
        lastscan timestamp(0) NOT NULL DEFAULT current_timestamp,
        enabled boolean NOT NULL DEFAULT TRUE,
        scantype varchar NOT NULL DEFAULT 'mailboxes',
        name varchar NOT NULL DEFAULT '',
        dimension varchar DEFAULT 'overworld',
        owner varchar NOT NULL DEFAULT '',
        sx int,
        sy int,
        sz int,
        fx int,
        fy int,
        fz int,
        PRIMARY KEY(scanrangeID)
    );
    CREATE TRIGGER onupdate BEFORE UPDATE ON postal_scan FOR EACH ROW EXECUTE PROCEDURE onupdate_changed();
    GRANT SELECT, INSERT, UPDATE ON postal_scan TO phoebot;

    GRANT SELECT, INSERT, UPDATE ON config TO phoebot;

COMMIT;
