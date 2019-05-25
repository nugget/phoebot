package main

import (
	"regexp"

	"github.com/bwmarrin/discordgo"
)

func regTemplate() (t Trigger) {
	t.Regexp = regexp.MustCompile("xyzzy")
	t.Hook = procTemplate
	t.Direct = true

	return t
}

func procTemplate(dm *discordgo.MessageCreate) error {
	// Uncomment these lines if you need to pull out substrings from
	// the original hook regular expression.
	//
	// t := regTemplate()
	//res := t.Regexp.FindStringSubmatch(dm.Content)
	//

	s.Dg.ChannelMessageSend(dm.ChannelID, "A hollow voice says 'plugh'")

	return nil
}
