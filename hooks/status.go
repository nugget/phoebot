package hooks

import (
	"fmt"
	"regexp"
	"strings"

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

func uname() string {
	u := strings.Builder{}

	u.WriteString(fmt.Sprintf("Phoebot/%s (%s)\n",
		builddata.VERSION,
		builddata.BUILDDATE,
	))

	u.WriteString(fmt.Sprintf("Built by %s@%s running %s\n",
		builddata.BUILDUSER,
		builddata.BUILDHOST,
		builddata.BUILDENV,
	))

	u.WriteString(fmt.Sprintf("Branch `%s` commit `%s`\n",
		builddata.GITBRANCH,
		builddata.GITCOMMIT,
	))

	if builddata.BUILDEMBEDLABEL != "" {
		u.WriteString(fmt.Sprintf("Label: %s\n",
			builddata.BUILDEMBEDLABEL,
		))
	}

	u.WriteString("Source code and issue tracker are at https://github.com/nugget/phoebot\n")

	return u.String()
}

func ProcStatus(s *state.State, dm *discordgo.MessageCreate) error {
	s.Dg.ChannelMessageSend(dm.ChannelID, uname())
	return nil
}
