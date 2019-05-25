package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func regStatus() (t Trigger) {
	t.Regexp = regexp.MustCompile("status report")
	t.Hook = procStatus
	t.Direct = true

	return t
}

func uname() string {
	u := strings.Builder{}

	u.WriteString(fmt.Sprintf("Phoebot/%s (%s)\n",
		VERSION,
		BUILDDATE,
	))

	u.WriteString(fmt.Sprintf("Built by %s@%s running %s\n",
		BUILDUSER,
		BUILDHOST,
		BUILDENV,
	))

	u.WriteString(fmt.Sprintf("Branch %s commit %s\n",
		GITBRANCH,
		GITCOMMIT,
	))

	if BUILDEMBEDLABEL != "" {
		u.WriteString(fmt.Sprintf("Label: %s\n",
			BUILDEMBEDLABEL,
		))
	}

	return u.String()
}

func procStatus(dm *discordgo.MessageCreate) error {
	s.Dg.ChannelMessageSend(dm.ChannelID, uname())
	return nil
}