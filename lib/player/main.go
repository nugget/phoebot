package player

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/nugget/phoebot/lib/db"
	"github.com/nugget/phoebot/lib/discord"
	"github.com/nugget/phoebot/lib/ipc"
	"github.com/nugget/phoebot/lib/phoelib"
	"github.com/nugget/phoebot/models"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type Player struct {
	Added         time.Time
	Changed       time.Time
	DiscordID     string // Discord player ID
	MinecraftUUID uuid.UUID
	MinecraftName string
	Email         string
	Ignored       bool
	Username      string //Discord user name
	Verified      bool
}

func PlayerFields() string {
	return "added, changed, playerid, minecraftname, username, verified"
}

func (p *Player) Scan(rows *sql.Rows) error {
	return rows.Scan(
		&p.Added,
		&p.Changed,
		&p.DiscordID,
		&p.MinecraftName,
		&p.Username,
		&p.Verified,
	)
}

func (p *Player) Parse() error {
	return nil
}

func (p *Player) SendMessage(message string) error {
	sentWhisper, sentDiscord := false, false

	if p.MinecraftName != "" {
		w := models.Whisper{p.MinecraftName, message}

		if ipc.ServerWhisperStream != nil {
			ipc.ServerWhisperStream <- w
			sentWhisper = true
		}
	}

	if p.DiscordID != "" {
		channel, err := discord.GetChannelByPlayerID(p.DiscordID)
		if err != nil {
			return err
		} else {
			discord.Session.ChannelMessageSend(channel.ID, message)
			sentDiscord = true
		}
	}

	logrus.WithFields(logrus.Fields{
		"Player":  p.MinecraftName,
		"whisper": sentWhisper,
		"discord": sentDiscord,
	}).Debug("Sent message to player")

	return nil
}

func GetPlayerFromMinecraftName(minecraftName string) (Player, error) {
	logrus.WithFields(logrus.Fields{
		"minecraftName": minecraftName,
	}).Trace("GetPlayerFromMinecratName")

	query := `SELECT ` + PlayerFields() + ` FROM player WHERE minecraftname = $1 ORDER BY added DESC LIMIT 1`

	phoelib.LogSQL(query, minecraftName)

	rows, err := db.DB.Query(query, minecraftName)
	if err != nil {
		return Player{}, err
	}
	defer rows.Close()

	p := Player{}

	for rows.Next() {
		err = p.Scan(rows)
		if err != nil {
			return Player{}, err
		}
		err = p.Parse()
		if err != nil {
			return Player{}, err
		}

		return p, nil
	}

	return p, fmt.Errorf("Not found")
}

func GetPlayerFromDiscordID(discordID string) (p Player, err error) {
	query := `SELECT ` + PlayerFields() + `FROM player WHERE playerid = $1 ORDER BY added DESC LIMIT 1`

	phoelib.LogSQL(query, discordID)
	rows, err := db.DB.Query(query, discordID)
	if err != nil {
		return Player{}, err
	}
	defer rows.Close()

	for rows.Next() {
		err := p.Scan(rows)
		if err != nil {
			return Player{}, err
		}
		err = p.Parse()
		if err != nil {
			return Player{}, err
		}
	}

	return p, nil
}

func UpdateFromDiscord(u *discordgo.User) error {
	logrus.WithField("u", u).Trace("player.Update")

	query := `INSERT INTO player (playerid, username)
			  SELECT $1, $2
			  ON CONFLICT (playerid) 
			     DO UPDATE SET username = $2
			        WHERE player.username <> $2`

	phoelib.LogSQL(query, u.ID, u.Username)
	_, err := db.DB.Exec(query, u.ID, u.Username)

	return err
}

func GameNickFromPlayerID(playerID string) (string, error) {
	logrus.Warn("Deprecated function call GameNickFromPlayerID")
	p, err := GetPlayerFromDiscordID(playerID)
	return p.MinecraftName, err
}

func PlayerIDFromGameNick(gameNick string) (string, error) {
	logrus.Warn("Deprecated function call PlayerIDFromGameNick")
	p, err := GetPlayerFromMinecraftName(gameNick)
	return p.DiscordID, err
}

func SendMessage(minecraftName, message string) error {
	p, err := GetPlayerFromMinecraftName(minecraftName)
	if err != nil {
		return err
	}
	return p.SendMessage(message)
}

func Advancement(playerName, advancement string) error {
	command := fmt.Sprintf("/advancement grant %s only %s", playerName, advancement)
	if ipc.ServerSayStream != nil {
		ipc.ServerSayStream <- command
	}
	return nil
}
