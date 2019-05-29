-- Deploy phoebot:ignore to pg

BEGIN;

    CREATE TABLE ignore (
        ignoreID uuid NOT NULL DEFAULT gen_random_uuid(),
        added timestamp(0) NOT NULL DEFAULT current_timestamp,
        changed timestamp(0) NOT NULL DEFAULT current_timestamp,
        deleted timestamp(0),
        enabled boolean NOT NULL DEFAULT TRUE,
        category varchar NOT NULL,
        target varchar NOT NULL,
        description varchar NOT NULL DEFAULT '',
        PRIMARY KEY(ignoreID)
    );
    CREATE TRIGGER onupdate BEFORE UPDATE ON ignore FOR EACH ROW EXECUTE PROCEDURE onupdate_changed();
    GRANT SELECT, INSERT, UPDATE ON ignore TO phoebot;

COMMIT;
