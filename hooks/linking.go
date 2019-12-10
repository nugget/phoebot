package hooks

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/nugget/phoebot/lib/db"
	"github.com/nugget/phoebot/lib/discord"
	"github.com/nugget/phoebot/lib/ipc"
	"github.com/nugget/phoebot/lib/phoelib"
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

	gameNick := res[1]
	if gameNick == "" {
		msg := fmt.Sprintf("Please tell me your in game player name: `!gamenick NICKNAME`")
		discord.Session.ChannelMessageSend(dm.ChannelID, msg)
		return fmt.Errorf("No gameNick supplied")
	}

	rand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")
	length := 5
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	code := strings.ToLower(b.String())

	query := `UPDATE player SET minecraftName = $2, verifycode = $3, verified = FALSE WHERE playerID = $1`
	phoelib.LogSQL(query, dm.Author.ID, gameNick, code)
	_, err := db.DB.Exec(query, dm.Author.ID, gameNick, code)
	if err != nil {
		discord.Session.ChannelMessageSend(dm.ChannelID, fmt.Sprintf("%v", err))
		return err
	}

	verifyMsg := fmt.Sprintf("Discord user %s says that they are you.  If this is correct, please use code '%s' on Discord to verify your identity.", dm.Author.Username, code)
	w := models.Whisper{gameNick, verifyMsg}
	if ipc.ServerWhisperStream != nil {
		ipc.ServerWhisperStream <- w

		msg := fmt.Sprintf("Please reply with `!verify CODE` using the code I just sent to %s in the game", gameNick)
		discord.Session.ChannelMessageSend(dm.ChannelID, msg)

		logrus.WithFields(logrus.Fields{
			"playerID":    dm.Author.ID,
			"verifyCode":  code,
			"gameNick":    gameNick,
			"discordName": dm.Author.Username,
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

	query := `UPDATE player SET verified = TRUE, verifyCode = '' WHERE verifyCode = $1 AND verified IS FALSE RETURNING playerID, minecraftName`

	phoelib.LogSQL(query, code)

	rows, err := db.DB.Query(query, code)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		playerID := "unknown"
		minecraftName := "unknown"

		err := rows.Scan(
			&playerID,
			&minecraftName,
		)
		if err != nil {
			return err
		}
		logrus.WithFields(logrus.Fields{
			"playerID":      playerID,
			"minecraftName": minecraftName,
		}).Info("Linked Minecraft Account")

		if minecraftName == "unknown" {

		}
		msg := fmt.Sprintf("You are successfully linked with Minecraft user %s", minecraftName)
		discord.Session.ChannelMessageSend(dm.ChannelID, msg)
	}

	return nil
}
