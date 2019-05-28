-- Deploy phoebot:mojangreleases to pg

BEGIN;

    CREATE TABLE mojangnews (
        articleID uuid NOT NULL DEFAULT gen_random_uuid(),
        added timestamp(0) NOT NULL DEFAULT current_timestamp,
        changed timestamp(0) NOT NULL DEFAULT current_timestamp,
        publishdate timestamp(0) NOT NULL DEFAULT current_timestamp,
        title varchar NOT NULL,
        url varchar NOT NULL,
        release boolean NOT NULL DEFAULT FALSE,
        product varchar NOT NULL DEFAULT '',
        version varchar NOT NULL DEFAULT '',
        PRIMARY KEY (articleID)
    );
    CREATE TRIGGER onupdate BEFORE UPDATE ON mojangnews FOR EACH ROW EXECUTE PROCEDURE onupdate_changed();
    GRANT SELECT, INSERT, UPDATE ON mojangnews TO phoebot;

    CREATE UNIQUE INDEX pubdates ON mojangnews(title, url, publishdate);

COMMIT;
