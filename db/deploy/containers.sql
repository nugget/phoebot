-- Deploy phoebot:containers to pg

BEGIN;
    CREATE TABLE container (
        containerID uuid NOT NULL DEFAULT gen_random_uuid(),
        added timestamp(0) NOT NULL DEFAULT current_timestamp,
        changed timestamp(0) NOT NULL DEFAULT current_timestamp,
        deleted timestamp(0),
        enabled boolean NOT NULL DEFAULT TRUE,
        id varchar NOT NULL DEFAULT '',
        name varchar NOT NULL DEFAULT '',
        playerID varchar NOT NULL DEFAULT '',
        dimension varchar DEFAULT 'overworld',
        x int,
        y int,
        z int,
        nbt varchar NOT NULL DEFAULT '{}',
        PRIMARY KEY(containerID)
    );
    CREATE TRIGGER onupdate BEFORE UPDATE ON container FOR EACH ROW EXECUTE PROCEDURE onupdate_changed();
    GRANT SELECT, INSERT, UPDATE ON container TO phoebot;

    ALTER TABLE player ADD COLUMN verifyCode varchar NOT NULL DEFAULT '';
    ALTER TABLE player ADD COLUMN verified bool NOT NULL DEFAULT FALSE;
    ALTER TABLE channel ADD COLUMN playerID varchar NOT NULL DEFAULT '';

COMMIT;
