package hooks

import (
	"regexp"

	"github.com/nugget/phoebot/lib/discord"

	"github.com/bwmarrin/discordgo"
)

func RegTemplate() (t Trigger) {
	t.Regexp = regexp.MustCompile("xyzzy")
	t.Hook = ProcTemplate
	t.Direct = true

	return t
}

func ProcTemplate(dm *discordgo.MessageCreate) error {
	// Uncomment these lines if you need to pull out substrings from
	// the original hook regular expression.
	//
	// t := RegTemplate()
	//res := t.Regexp.FindStringSubmatch(dm.Content)
	//

	discord.Session.ChannelMessageSend(dm.ChannelID, "A hollow voice says 'plugh'")

	return nil
}
