-- Deploy phoebot:initdb to pg

BEGIN;

    CREATE EXTENSION pgcrypto;

    CREATE TABLE config (
        added timestamp(0) NOT NULL DEFAULT current_timestamp,
        changed timestamp(0) NOT NULL DEFAULT current_timestamp,
        key varchar NOT NULL,
        value varchar,
        PRIMARY KEY(key)
    );

    GRANT SELECT, UPDATE ON config TO phoebot;


    CREATE TABLE user (
        added timestamp(0) NOT NULL DEFAULT current_timestamp,
        changed timestamp(0) NOT NULL DEFAULT current_timestamp,
        userID varchar NOT NULL,
        minecraftUUID uuid,
        timezone varchar,
        ignored boolean NOT NULL DEFAULT FALSE,
        PRIMARY KEY(userID)
    );

    GRANT SELECT, INSERT, UPDATE ON user TO phoebot;


    CREATE TABLE acl (
        added timestamp(0) NOT NULL DEFAULT current_timestamp,
        changed timestamp(0) NOT NULL DEFAULT current_timestamp,
        deleted timestamp(0),
        userID varchar NOT NULL REFERENCES user.userID,
        key varchar NOT NULL,
        PRIMARY KEY(userID, key)
    );

    GRANT SELECT, INSERT, UPDATE ON user TO phoebot;


    CREATE TABLE subscription (
        added timestamp(0) NOT NULL DEFAULT current_timestamp,
        changed timestamp(0) NOT NULL DEFAULT current_timestamp,
        deleted timestamp(0),
        subscriptionID uuid NOT NULL DEFAULT gen_random_uuid(),
        channelID varchar NOT NULL,
        class varchar NOT NULL,
        name varchar NOT NULL,
        target varchar,
        PRIMARY KEY(subscriptionID)
    );

    GRANT SELECT, INSERT, UPDATE ON subscription TO phoebot;


    CREATE TABLE product (
        added timestamp(0) NOT NULL DEFAULT current_timestamp,
        changed timestamp(0) NOT NULL DEFAULT current_timestamp,
        deleted timestamp(0),
        class varchar NOT NULL,
        name varchar NOT NULL,
        version varchar NOT NULL,
        PRIMARY KEY(class, name)
    )

    GRANT SELECT, INSERT, UPDATE ON product TO phoebot;

COMMIT;
