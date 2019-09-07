-- Deploy phoebot:poi-extras to pg

BEGIN;

    ALTER TABLE poi ADD COLUMN dimension varchar DEFAULT 'overworld';
    ALTER TABLE poi ADD COLUMN private boolean NOT NULL DEFAULT FALSE;

COMMIT;
