package hooks

import (
	"regexp"

	"github.com/nugget/phoebot/lib/builddata"
	"github.com/nugget/phoebot/lib/discord"

	"github.com/bwmarrin/discordgo"
)

func RegStatus() (t Trigger) {
	t.Regexp = regexp.MustCompile("!version")
	t.Hook = ProcStatus
	t.Direct = false

	return t
}

func ProcStatus(dm *discordgo.MessageCreate) error {
	discord.Session.ChannelMessageSend(dm.ChannelID, builddata.Uname())
	return nil
}
