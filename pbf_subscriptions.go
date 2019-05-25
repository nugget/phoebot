package main

import (
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
