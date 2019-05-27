package hooks

import (
	"regexp"

	"github.com/bwmarrin/discordgo"
)

type HookFunction func(*discordgo.MessageCreate) error

type Trigger struct {
	Regexp *regexp.Regexp
	Hook   HookFunction
	Direct bool
}
