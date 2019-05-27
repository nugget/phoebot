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


    CREATE TABLE discordlog (
        added timestamp(0) NOT NULL DEFAULT current_timestamp,
        changed timestamp(0) NOT NULL DEFAULT current_timestamp,
        messageID varchar NOT NULL,
        channelID varchar NOT NULL,
        guildID varchar NOT NULL,
        playerID varchar NOT NULL,
        messageType int NOT NULL,
        content varchar NOT NULL,
        PRIMARY KEY(messageID)
    );
    CREATE TRIGGER onupdate BEFORE UPDATE ON discordlog FOR EACH ROW EXECUTE PROCEDURE onupdate_changed();
    GRANT SELECT, INSERT ON discordlog TO phoebot;

COMMIT;
