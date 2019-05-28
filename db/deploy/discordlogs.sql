-- Deploy phoebot:discordlogs to pg

BEGIN;

    CREATE TABLE channel (
        added timestamp(0) NOT NULL DEFAULT current_timestamp,
        changed timestamp(0) NOT NULL DEFAULT current_timestamp,
        channelID varchar NOT NULL,
        guildID varchar NOT NULL,
        name varchar NOT NULL,
        channelType int NOT NULL,
        topic varchar NOT NULL,
        PRIMARY KEY(channelID)
    );
    CREATE TRIGGER onupdate BEFORE UPDATE ON channel FOR EACH ROW EXECUTE PROCEDURE onupdate_changed();
    GRANT SELECT, INSERT, UPDATE ON channel TO phoebot;

    CREATE TABLE lastseen (
        added timestamp(0) NOT NULL DEFAULT current_timestamp,
        changed timestamp(0) NOT NULL DEFAULT current_timestamp,
        playerID varchar NOT NULL REFERENCES player(playerID),
        guildID varchar NOT NULL,
        channelID varchar NOT NULL REFERENCES channel(channelID),
        content varchar NOT NULL,
        PRIMARY KEY(playerID, guildID)
    );
    CREATE TRIGGER onupdate BEFORE UPDATE ON lastseen FOR EACH ROW EXECUTE PROCEDURE onupdate_changed();
    GRANT SELECT, INSERT, UPDATE ON lastseen TO phoebot;

COMMIT;
