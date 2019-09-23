-- Deploy phoebot:mapowner to pg

BEGIN;

    ALTER TABLE map ADD COLUMN owner varchar NOT NULL DEFAULT 'MacNugget';
    ALTER TABLE poi ADD COLUMN owner varchar NOT NULL DEFAULT 'MacNugget';

COMMIT;
