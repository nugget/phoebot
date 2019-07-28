-- Deploy phoebot:map to pg

BEGIN;

    CREATE TABLE map (
        mapID int NOT NULL,
        added timestamp(0) NOT NULL DEFAULT current_timestamp,
        changed timestamp(0) NOT NULL DEFAULT current_timestamp,
        deleted timestamp(0),
        scale int NOT NULL DEFAULT 0,
        lX int NOT NULL,
        lZ int NOT NULL,
        rX int NOT NULL,
        rZ int NOT NULL,
        PRIMARY KEY(mapID)
    );
    CREATE TRIGGER onupdate BEFORE UPDATE ON map FOR EACH ROW EXECUTE PROCEDURE onupdate_changed();
    GRANT SELECT, INSERT, UPDATE ON map TO phoebot;

    CREATE TABLE poi (
        poiID uuid NOT NULL DEFAULT gen_random_uuid(),
        added timestamp(0) NOT NULL DEFAULT current_timestamp,
        changed timestamp(0) NOT NULL DEFAULT current_timestamp,
        deleted timestamp(0),
        class varchar NOT NULL,
        x int NOT NULL,
        y int NOT NULL,
        z int NOT NULL,
        description varchar NOT NULL,
        PRIMARY KEY(poiID)
    );

    CREATE TRIGGER onupdate BEFORE UPDATE ON poi FOR EACH ROW EXECUTE PROCEDURE onupdate_changed();
    GRANT SELECT, INSERT, UPDATE ON poi TO phoebot;

COMMIT;
