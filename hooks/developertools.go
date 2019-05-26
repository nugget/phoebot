package hooks

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/nugget/phoebot"
	"github.com/nugget/phoebot/models"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

func RegLoglevel() (t models.Trigger) {
	t.Regexp = regexp.MustCompile("(?i)set loglevel to ([A-Z]+)")
	t.Hook = ProcLoglevel
	t.Direct = true

	return t
}

func ProcLoglevel(dm *discordgo.MessageCreate) error {
	t := RegLoglevel()
	res := t.Regexp.FindStringSubmatch(dm.Content)

	phoebot.DebugSlice(res)

	reqLevel := strings.ToLower(res[1])

	oldLevel := logrus.GetLevel()

	newLevel, err := phoebot.LogLevel(reqLevel)
	if err != nil {
		logrus.WithError(err).Info("Failed to set LogLevel")
		main.PB.Dg.ChannelMessageSend(dm.ChannelID, fmt.Sprintf("%s", err))
		return err
	}

	if oldLevel != newLevel {
		logrus.WithFields(logrus.Fields{
			"old": oldLevel,
			"new": newLevel,
		}).Info("Console loging level changed")
	}

	message := fmt.Sprintf("Console logging level is '%s'", reqLevel)
	main.PB.Dg.ChannelMessageSend(dm.ChannelID, message)

	return nil
}
