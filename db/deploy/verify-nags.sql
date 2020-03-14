-- Deploy phoebot:verify-nags to pg

BEGIN;

    ALTER TABLE player DROP COLUMN timezone;
    ALTER TABLE player DROP COLUMN locale;
    ALTER TABLE player DROP COLUMN ignored;
    ALTER TABLE player DROP COLUMN verifycode;

    CREATE TABLE verify (
        added timestamp(0) NOT NULL DEFAULT current_timestamp,
        deleted timestamp(0),
        minecraftname varchar NOT NULL DEFAULT '',
        code varchar NOT NULL DEFAULT '',
        lastnag timestamp(0) without time zone NOT NULL DEFAULT current_timestamp,
        PRIMARY KEY(minecraftname)
    );

    CREATE TRIGGER onupdate BEFORE UPDATE ON verify FOR EACH ROW EXECUTE PROCEDURE onupdate_changed();
    GRANT SELECT, INSERT, UPDATE ON verify TO phoebot;

COMMIT;
