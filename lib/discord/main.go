package discord

import (
	"fmt"

	"github.com/nugget/phoebot/lib/db"
	"github.com/nugget/phoebot/lib/phoelib"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

var (
	Session *discordgo.Session
)

func RecordLog(m *discordgo.MessageCreate) error {
	if m.GuildID == "" || m.ChannelID == "" || m.Type > 0 {
		logrus.WithFields(logrus.Fields{
			"playerID":   m.Author.ID,
			"authorName": m.Author.Username,
			"guildID":    m.GuildID,
			"channelID":  m.ChannelID,
			"type":       m.Type,
		}).Trace("Not saving private message")
		return nil
	}

	query := `INSERT INTO lastseen (playerID, guildID, channelID, content)
			  SELECT $1, $2, $3, $4
			  ON CONFLICT (playerID, guildID)
			    DO UPDATE SET channelID = $3, content = $4`

	phoelib.LogSQL(query, m.Author.ID, m.GuildID, m.ChannelID, m.Content)
	_, err := db.DB.Exec(query, m.Author.ID, m.GuildID, m.ChannelID, m.Content)
	return err
}

func UpdateChannel(c *discordgo.Channel) error {
	query := `INSERT INTO channel (channelid, guildid, name, channeltype, topic)
			  SELECT $1, $2, $3, $4, $5
			  ON CONFLICT (channelid) 
			     DO UPDATE SET name = $3, topic = $5
			        WHERE channel.name <> $3 OR channel.topic <> $5`

	channelName := c.Name

	switch c.Type {
	case 0:
		// This is a regular channel
		channelName = fmt.Sprintf("#%s", c.Name)
	case 1:
		// This is PM channel
		channelName = fmt.Sprintf("@%s", c.Recipients[0])
		fmt.Printf("---\n%+v\n---\n", c)
	default:
		logrus.WithFields(logrus.Fields{
			"type":    c.Type,
			"id":      c.ID,
			"guildID": c.GuildID,
			"name":    c.Name,
		}).Info("Unknown channel type")
	}

	phoelib.LogSQL(query, c.ID, c.GuildID, channelName, string(c.Type), c.Topic)
	_, err := db.DB.Exec(query, c.ID, c.GuildID, channelName, c.Type, c.Topic)
	return err
}

func SetPMOwner(m *discordgo.MessageCreate) error {
	query := `UPDATE channel SET playerID = $2 WHERE channelID = $1 AND playerID <> $2 RETURNING channelID`

	_, err := db.DB.Exec(query, m.ChannelID, m.Author.ID)
	return err
}

func GetChannel(id string) (*discordgo.Channel, error) {
	channel, err := Session.Channel(id)
	if err != nil {
		return channel, err
	}

	err = UpdateChannel(channel)
	return channel, err
}

func GetChannelByName(name string) (*discordgo.Channel, error) {
	query := `SELECT channelid FROM channel WHERE name ILIKE $1 or name ILIKE '#'||$1`

	phoelib.LogSQL(query, name)
	rows, err := db.DB.Query(query, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var channelID string

		err := rows.Scan(&channelID)
		if err != nil {
			return nil, err
		}

		return GetChannel(channelID)
	}

	return nil, fmt.Errorf("Discord channel not found for name %s", name)
}

func GetChannelByPlayerID(playerID string) (*discordgo.Channel, error) {
	query := `SELECT channelid FROM channel WHERE playerID = $1`

	phoelib.LogSQL(query, playerID)
	rows, err := db.DB.Query(query, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var channelID string

		err := rows.Scan(&channelID)
		if err != nil {
			return nil, err
		}

		fmt.Println("Looking for channelID", channelID)
		return GetChannel(channelID)
	}

	return nil, fmt.Errorf("Discord channel not found for playerID %s", playerID)
}
