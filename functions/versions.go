package main

import (
	"fmt"
	"regexp"

	"github.com/nugget/phoebot/models"

	"github.com/blang/semver"
	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

func RegVersion() (t models.Trigger) {
	t.Regexp = regexp.MustCompile("version report")
	t.Hook = procVersion
	t.Direct = true

	return t
}

func ProcVersion(dm *discordgo.MessageCreate) error {
	cutoff := semver.MustParse("0.0.0")
	logrus.WithField("cutoff", cutoff).Debug("Ignoring products older than this")

	mS := discordgo.MessageSend{}

	mE := discordgo.MessageEmbed{}
	mE.Description = "Current Minecraft Versions:"

	mE.Fields = make([]*discordgo.MessageEmbedField, 0)

	for _, p := range s.Products {
		if p.Latest.Version.GT(cutoff) {
			mE.Fields = append(mE.Fields, &discordgo.MessageEmbedField{
				Name:   fmt.Sprintf("%s %s", p.Class, p.Name),
				Value:  fmt.Sprintf("%s", p.Latest.Version),
				Inline: true,
			})
		} else {
			logrus.WithFields(logrus.Fields{
				"class":   p.Class,
				"name":    p.Name,
				"version": p.Latest.Version,
				"time":    p.Latest.Time,
			}).Debug("Skipped product")
		}

	}

	mS.Embed = &mE

	if len(mE.Fields) > 0 {
		s.Dg.ChannelMessageSendComplex(dm.ChannelID, &mS)
	} else {
		s.Dg.ChannelMessageSend(dm.ChannelID, "I haven't seen any new versions lately, sorry. Try again later.")
	}
	return nil
}
