package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/nugget/phoebot/lib/state"
	"github.com/nugget/phoebot/models"

	"github.com/bwmarrin/discordgo"
)

func RegStatus() (t models.Trigger) {
	t.Regexp = regexp.MustCompile("status report")
	t.Hook = procStatus
	t.Direct = true

	return t
}

func uname() string {
	u := strings.Builder{}

	u.WriteString(fmt.Sprintf("Phoebot/%s (%s)\n",
		state.VERSION,
		state.BUILDDATE,
	))

	u.WriteString(fmt.Sprintf("Built by %s@%s running %s\n",
		state.BUILDUSER,
		state.BUILDHOST,
		state.BUILDENV,
	))

	u.WriteString(fmt.Sprintf("Branch `%s` commit `%s`\n",
		state.GITBRANCH,
		state.GITCOMMIT,
	))

	if state.BUILDEMBEDLABEL != "" {
		u.WriteString(fmt.Sprintf("Label: %s\n",
			state.BUILDEMBEDLABEL,
		))
	}

	u.WriteString("Source code and issue tracker are at https://github.com/nugget/phoebot\n")

	return u.String()
}

func ProcStatus(dm *discordgo.MessageCreate) error {
	s.Dg.ChannelMessageSend(dm.ChannelID, uname())
	return nil
}
