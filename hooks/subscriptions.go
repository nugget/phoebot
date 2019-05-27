package hooks

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/nugget/phoebot/lib/discord"
	"github.com/nugget/phoebot/lib/ipc"
	"github.com/nugget/phoebot/lib/phoelib"
	"github.com/nugget/phoebot/lib/products"
	"github.com/nugget/phoebot/lib/subscriptions"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

func RegSubscriptions() (t Trigger) {
	t.Regexp = regexp.MustCompile("(?i)((un)?(sub)(scribe)?) ([^ ]+) ([^ ]+) ?(.*)")
	t.Hook = ProcSubscriptions
	t.Direct = true

	return t
}

func ProcSubscriptions(dm *discordgo.MessageCreate) error {
	t := RegSubscriptions()
	res := t.Regexp.FindStringSubmatch(dm.Content)

	phoelib.DebugSlice(res)

	if len(res) == 8 {
		var err error

		sc := ipc.SubscriptionChannel{}

		xUN := strings.ToLower(res[2])
		xSUB := strings.ToLower(res[3])

		class := res[5]
		name := res[6]

		p, err := products.GetProduct(class, name)
		if err != nil {
			logrus.WithError(err).Info("Unable to get matching product")
			discord.Session.ChannelMessageSend(dm.ChannelID, "I've never heard of that one, sorry.")
			return nil
		}

		sc.Sub.ChannelID = dm.ChannelID
		sc.Sub.Class = p.Class
		sc.Sub.Name = p.Name
		sc.Sub.Target = res[7]

		if xUN == "un" {
			sc.Operation = "DROP"
		} else if xSUB == "sub" {
			sc.Operation = "ADD"
		}

		ipc.SubStream <- sc
	}
	return nil
}

func RegListSubscriptions() (t Trigger) {
	t.Regexp = regexp.MustCompile("(?i)list subs")
	t.Hook = ProcListSubscriptions
	t.Direct = true

	return t
}

func ProcListSubscriptions(dm *discordgo.MessageCreate) error {
	logrus.WithFields(logrus.Fields{
		"channel": dm.ChannelID,
	}).Info("Listing subscriptions")

	localSubs, err := subscriptions.GetByChannel(dm.ChannelID)
	if err != nil {
		logrus.WithError(err).Error("GetByChannel failed")
		return err
	}

	subCount := len(localSubs)
	logrus.WithField("subCount", subCount).Info("Active subscription count")

	if subCount == 0 {
		discord.Session.ChannelMessageSend(dm.ChannelID, "I don't have any subscriptions for this channel")
		return nil
	}

	u := strings.Builder{}

	u.WriteString("This channel is set up to receive announcements for:\n")

	for _, v := range localSubs {
		subDesc := fmt.Sprintf(" * %s %s", v.Class, v.Name)

		cleanTarget := strings.Replace(v.Target, "@", "", -1)

		if cleanTarget != "" {
			subDesc = fmt.Sprintf("%s (to %s)", subDesc, cleanTarget)
		}

		u.WriteString(fmt.Sprintf("%s\n", subDesc))
	}

	discord.Session.ChannelMessageSend(dm.ChannelID, u.String())

	return nil
}
