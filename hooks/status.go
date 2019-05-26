package hooks

import (
	"regexp"

	"github.com/nugget/phoebot/lib/builddata"
	"github.com/nugget/phoebot/lib/state"

	"github.com/bwmarrin/discordgo"
)

func RegStatus() (t Trigger) {
	t.Regexp = regexp.MustCompile("status report")
	t.Hook = ProcStatus
	t.Direct = true

	return t
}

func ProcStatus(s *state.State, dm *discordgo.MessageCreate) error {
	s.Dg.ChannelMessageSend(dm.ChannelID, builddata.Uname())
	return nil
}
