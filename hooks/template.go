package hooks

import (
	"regexp"

	"github.com/nugget/phoebot/lib/state"

	"github.com/bwmarrin/discordgo"
)

func RegTemplate() (t Trigger) {
	t.Regexp = regexp.MustCompile("xyzzy")
	t.Hook = ProcTemplate
	t.Direct = true

	return t
}

func ProcTemplate(s *state.State, dm *discordgo.MessageCreate) error {
	// Uncomment these lines if you need to pull out substrings from
	// the original hook regular expression.
	//
	// t := RegTemplate()
	//res := t.Regexp.FindStringSubmatch(dm.Content)
	//

	s.Dg.ChannelMessageSend(dm.ChannelID, "A hollow voice says 'plugh'")

	return nil
}
