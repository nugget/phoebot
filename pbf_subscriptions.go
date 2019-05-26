package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/nugget/phoebot/models"
	"github.com/sirupsen/logrus"

	"github.com/bwmarrin/discordgo"
)

func regSubscriptions() (t Trigger) {
	t.Regexp = regexp.MustCompile("(?i)((un)?(sub)(scribe)?) ([^ ]+) ([^ ]+) ?(.*)")
	t.Hook = procSubscriptions
	t.Direct = true

	return t
}

func procSubscriptions(dm *discordgo.MessageCreate) error {
	t := regSubscriptions()
	res := t.Regexp.FindStringSubmatch(dm.Content)

	if len(res) == 8 {
		var err error

		sc := models.SubChannel{}

		xUN := strings.ToLower(res[2])
		xSUB := strings.ToLower(res[3])

		class := res[5]
		name := res[6]

		p, err := s.GetProduct(class, name)
		if err != nil {
			logrus.WithError(err).Warn("Unable to get product")
			s.Dg.ChannelMessageSend(dm.ChannelID, "I've never heard of that one, sorry.")
		} else {
			sc.Sub.ChannelID = dm.ChannelID
			sc.Sub.Class = p.Class
			sc.Sub.Name = p.Name
			sc.Sub.Target = res[7]

			if xUN == "un" {
				sc.Operation = "DROP"
			} else if xSUB == "sub" {
				sc.Operation = "ADD"
			}

			subStream <- sc
		}
	}
	return nil
}

func regListSubscriptions() (t Trigger) {
	t.Regexp = regexp.MustCompile("(?i)list subs")
	t.Hook = procListSubscriptions
	t.Direct = true

	return t
}

func procListSubscriptions(dm *discordgo.MessageCreate) error {
	var localSubs []models.Subscription

	for _, v := range s.Subscriptions {
		if v.ChannelID == dm.ChannelID {
			localSubs = append(localSubs, v)
		}
	}

	subCount := len(localSubs)
	logrus.WithField("subCount", subCount).Info("Active subscription count")

	if subCount == 0 {
		s.Dg.ChannelMessageSend(dm.ChannelID, "I don't have any subscriptions for this channel")
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

	s.Dg.ChannelMessageSend(dm.ChannelID, u.String())

	return nil
}
