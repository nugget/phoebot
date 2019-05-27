-- Deploy phoebot:function/onupdate to pg

BEGIN;

    CREATE OR REPLACE FUNCTION onupdate_changed() RETURNS trigger AS $$
        BEGIN
            NEW.changed := (current_timestamp AT TIME ZONE 'UTC');
            RETURN NEW;
        END;
        $$ LANGUAGE plpgsql;

COMMIT;
