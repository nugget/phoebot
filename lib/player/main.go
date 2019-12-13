package player

import (
	"github.com/nugget/phoebot/lib/db"
	"github.com/nugget/phoebot/lib/discord"
	"github.com/nugget/phoebot/lib/ipc"
	"github.com/nugget/phoebot/lib/phoelib"
	"github.com/nugget/phoebot/models"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

func UpdateFromDiscord(u *discordgo.User) error {
	logrus.WithField("u", u).Trace("player.Update")

	query := `INSERT INTO player (playerid, username, locale)
			  SELECT $1, $2, $3
			  ON CONFLICT (playerid) 
			     DO UPDATE SET username = $2, locale = $3
			        WHERE player.username <> $2 OR player.locale <> $3`

	phoelib.LogSQL(query, u.ID, u.Username, u.Locale)
	_, err := db.DB.Exec(query, u.ID, u.Username, u.Locale)

	return err
}

func GameNickFromPlayerID(playerID string) (string, error) {
	query := `SELECT minecraftName FROM player WHERE playerID = $1 AND verified IS TRUE ORDER BY changed DESC LIMIT 1`

	phoelib.LogSQL(query, playerID)

	rows, err := db.DB.Query(query, playerID)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	gameNick := ""

	for rows.Next() {
		err := rows.Scan(&gameNick)
		if err != nil {
			return "", err
		}
	}

	return gameNick, nil
}

func PlayerIDFromGameNick(gameNick string) (string, error) {
	query := `SELECT playerID FROM player WHERE minecraftName = $1 AND verified IS TRUE ORDER BY changed DESC LIMIT 1`

	phoelib.LogSQL(query, gameNick)

	rows, err := db.DB.Query(query, gameNick)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	playerID := ""

	for rows.Next() {
		err := rows.Scan(&playerID)
		if err != nil {
			return "", err
		}
	}

	return playerID, nil
}

func SendMessage(playerName, message string) error {
	sentWhisper, sentDiscord := false, false

	if playerName != "" {
		w := models.Whisper{Who: playerName, Message: message}

		if ipc.ServerWhisperStream != nil {
			ipc.ServerWhisperStream <- w
			sentWhisper = true
		}
	}

	playerID, err := PlayerIDFromGameNick(playerName)
	if err != nil {
		return err
	}

	if playerID != "" {
		channel, err := discord.GetChannelByPlayerID(playerID)
		if err != nil {
			return err
		} else {
			discord.Session.ChannelMessageSend(channel.ID, message)
			sentDiscord = true
		}
	}

	logrus.WithFields(logrus.Fields{
		"Player":  playerName,
		"whisper": sentWhisper,
		"discord": sentDiscord,
	}).Debug("Sent message to player")

	return nil
}
