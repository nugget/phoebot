package hooks

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/nugget/phoebot/lib/console"
	"github.com/nugget/phoebot/lib/discord"
	"github.com/sirupsen/logrus"

	"github.com/bwmarrin/discordgo"
)

func RegServerInfo() (t Trigger) {
	t.Regexp = regexp.MustCompile("(?i)server info")
	t.Hook = ProcServerInfo
	t.Direct = false

	return t
}

func ProcServerInfo(dm *discordgo.MessageCreate) error {
	si, err := console.GetServerInfo()
	if err != nil {
		logrus.WithError(err).Error("ProcServerInfo failure")
		return err
	}

	resp := fmt.Sprintf(
		"%s\nTPS from last 1m, 5m, 15m: %2.2f, %2.2f, %2.2f\n%d/%d players online",
		si.Version,
		si.Tps1,
		si.Tps5,
		si.Tps15,
		len(si.Players),
		si.MaxPlayers,
	)

	discord.Session.ChannelMessageSend(dm.ChannelID, resp)
	return nil
}

func RegServerList() (t Trigger) {
	t.Regexp = regexp.MustCompile("(?i)who is on the server")
	t.Hook = ProcServerList
	t.Direct = false

	return t
}

func ProcServerList(dm *discordgo.MessageCreate) error {
	si, err := console.GetServerInfo()
	if err != nil {
		logrus.WithError(err).Error("ProcServerList failure")
		return err
	}

	resp := fmt.Sprintf(
		"%d/%d players online:\n> %s",
		len(si.Players),
		si.MaxPlayers,
		strings.Join(si.Players, ", "),
	)

	discord.Session.ChannelMessageSend(dm.ChannelID, resp)
	return nil
}
