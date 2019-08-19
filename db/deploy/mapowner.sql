-- Deploy phoebot:mapowner to pg

BEGIN;

    ALTER TABLE map ADD COLUMN owner varchar NOT NULL DEFAULT 'MacNugget';
    ALTER TABLE poi ADD COLUMN owner varchar NOT NULL DEFAULT 'MacNugget';

    CREATE UNIQUE INDEX poi_index ON poi(class, description);

COMMIT;
