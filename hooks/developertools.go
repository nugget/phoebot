package hooks

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/nugget/phoebot/lib/discord"
	"github.com/nugget/phoebot/lib/phoelib"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

func RegLoglevel() (t Trigger) {
	t.Regexp = regexp.MustCompile("(?i)set loglevel to ([A-Z]+)")
	t.Hook = ProcLoglevel
	t.Direct = true

	return t
}

func ProcLoglevel(dm *discordgo.MessageCreate) error {
	t := RegLoglevel()
	res := t.Regexp.FindStringSubmatch(dm.Content)

	phoelib.DebugSlice(res)

	reqLevel := strings.ToLower(res[1])

	oldLevel := logrus.GetLevel()

	newLevel, err := phoelib.LogLevel(reqLevel)
	if err != nil {
		logrus.WithError(err).Info("Failed to set LogLevel")
		discord.Session.ChannelMessageSend(dm.ChannelID, fmt.Sprintf("%s", err))
		return err
	}

	if oldLevel != newLevel {
		logrus.WithFields(logrus.Fields{
			"old": oldLevel,
			"new": newLevel,
		}).Info("Console loging level changed")
	}

	message := fmt.Sprintf("Console logging level is '%s'", reqLevel)
	discord.Session.ChannelMessageSend(dm.ChannelID, message)

	return nil
}
