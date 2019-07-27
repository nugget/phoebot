package hooks

import (
	"regexp"

	"github.com/bwmarrin/discordgo"
)

type HookFunction func(*discordgo.MessageCreate) error
type GameHookFunction func(string) (string, error)

type Trigger struct {
	Regexp   *regexp.Regexp
	Hook     HookFunction
	GameHook GameHookFunction
	Direct   bool
	InGame   bool
	ACL      string
}
