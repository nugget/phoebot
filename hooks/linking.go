package hooks

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/nugget/phoebot/lib/db"
	"github.com/nugget/phoebot/lib/discord"
	"github.com/nugget/phoebot/lib/ipc"
	"github.com/nugget/phoebot/lib/phoelib"
	"github.com/nugget/phoebot/lib/player"
	"github.com/nugget/phoebot/models"
	"github.com/sirupsen/logrus"

	"github.com/bwmarrin/discordgo"
)

func RegLinkRequest() (t Trigger) {
	t.Regexp = regexp.MustCompile("(?i)!gamenick (.*)")
	t.Hook = ProcLinkRequest
	t.Direct = true

	return t
}

func ProcLinkRequest(dm *discordgo.MessageCreate) error {
	t := RegLinkRequest()
	res := t.Regexp.FindStringSubmatch(dm.Content)

	minecraftName := res[1]
	if minecraftName == "" {
		msg := fmt.Sprintf("Please tell me your in game player name: `!gamenick NICKNAME`")
		discord.Session.ChannelMessageSend(dm.ChannelID, msg)
		return fmt.Errorf("No gameNick supplied")
	}

	code, _, err := player.GenerateCode(minecraftName)
	if err != nil {
		discord.Session.ChannelMessageSend(dm.ChannelID, fmt.Sprintf("%v", err))
		logrus.WithError(err).Error("ProcLinkRequest GenerateCode")
	}

	verifyMsg := fmt.Sprintf("Discord user %s says that they are you.  If this is correct, please use code '%s' on Discord to verify your identity.", dm.Author.Username, code)
	w := models.Whisper{minecraftName, verifyMsg}
	if ipc.ServerWhisperStream != nil {
		ipc.ServerWhisperStream <- w

		msg := fmt.Sprintf("Please reply with `!verify CODE` using the code I just sent to %s in the game", minecraftName)
		discord.Session.ChannelMessageSend(dm.ChannelID, msg)

		logrus.WithFields(logrus.Fields{
			"playerID":      dm.Author.ID,
			"verifyCode":    code,
			"minecraftName": minecraftName,
			"discordName":   dm.Author.Username,
		}).Info("Sent verification request in game")

	}

	return nil
}

func RegLinkVerify() (t Trigger) {
	t.Regexp = regexp.MustCompile("(?i)!verify (.*)")
	t.Hook = ProcLinkVerify
	t.Direct = true

	return t
}

func ProcLinkVerify(dm *discordgo.MessageCreate) error {
	t := RegLinkVerify()
	res := t.Regexp.FindStringSubmatch(dm.Content)

	code := strings.ToLower(res[1])

	minecraftName, _, err := player.LookupCode(code)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"User": dm.Author.Username,
			"code": code,
		}).Warn("Unrecognized verify code")

		// msg := "I don't recognize that code, sorry."
		// discord.Session.ChannelMessageSend(dm.ChannelID, msg)
		return nil
	}

	query := `UPDATE player SET minecraftname = $2, verified = TRUE WHERE playerID = $1 AND verified IS FALSE RETURNING username`

	phoelib.LogSQL(query, dm.Author.ID, minecraftName)
	rows, err := db.DB.Query(query, dm.Author.ID, minecraftName)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		logrus.WithFields(logrus.Fields{
			"username":      dm.Author.Username,
			"minecraftName": minecraftName,
		}).Info("Linked Minecraft Account")

		err = player.Advancement(minecraftName, "phoenixcraft:phoenixcraft/discord")
		if err != nil {
			logrus.WithError(err).Warn("Unable to grant advancement")
		}

		msg := fmt.Sprintf("You are successfully linked with Minecraft user %s", minecraftName)
		discord.Session.ChannelMessageSend(dm.ChannelID, msg)
	}

	query = `UPDATE verify SET deleted = current_timestamp at time zone 'UTC' WHERE code = $1`
	_, err = db.DB.Exec(query, code)
	if err != nil {
		logrus.WithError(err).Error("Unable to delete used verify code")
		return err
	}

	return nil
}
