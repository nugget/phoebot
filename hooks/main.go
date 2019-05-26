package hooks

import (
	"regexp"

	"github.com/nugget/phoebot/lib/state"

	"github.com/bwmarrin/discordgo"
)

type HookFunction func(*state.State, *discordgo.MessageCreate) error

type Trigger struct {
	Regexp *regexp.Regexp
	Hook   HookFunction
	Direct bool
}
