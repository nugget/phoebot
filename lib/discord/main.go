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
	query := `INSERT INTO discordlog (messageid, channelid, guildid, playerid, messagetype, content)
			  SELECT $1, $2, $3, $4, $5, $6`

	phoelib.LogSQL(query, m.ID, m.ChannelID, m.GuildID, m.Author.ID, string(m.Type), m.Content)
	_, err := db.DB.Exec(query, m.ID, m.ChannelID, m.GuildID, m.Author.ID, m.Type, m.Content)
	return err
}

func UpdateChannel(c *discordgo.Channel) error {
	query := `INSERT INTO channel (channelid, guildid, name, channeltype, topic)
			  SELECT $1, $2, $3, $4, $5
			  ON CONFLICT (channelid) 
			     DO UPDATE SET name = $3, topic = $5
			        WHERE channel.name <> $3 OR channel.topic <> $5`

	switch c.Type {
	case 0:
		// This is a regular channel
		c.Name = fmt.Sprintf("#%s", c.Name)
	case 1:
		// This is PM channel
		c.Name = fmt.Sprintf("@%s", c.Recipients[0])
	default:
		logrus.WithFields(logrus.Fields{
			"type":    c.Type,
			"id":      c.ID,
			"guildID": c.GuildID,
			"name":    c.Name,
		}).Info("Unknown channel type")
	}

	phoelib.LogSQL(query, c.ID, c.GuildID, c.Name, string(c.Type), c.Topic)
	_, err := db.DB.Exec(query, c.ID, c.GuildID, c.Name, c.Type, c.Topic)
	return err
}

func GetChannel(id string) (*discordgo.Channel, error) {
	channel, err := Session.State.Channel(id)
	if err != nil {
		return channel, err
	}

	err = UpdateChannel(channel)
	return channel, err
}
