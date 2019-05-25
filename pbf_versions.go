package main

import (
	"fmt"
	"regexp"

	"github.com/blang/semver"
	"github.com/bwmarrin/discordgo"
)

func regVersion() (t Trigger) {
	t.Regexp = regexp.MustCompile("version report")
	t.Hook = procVersion
	t.Direct = true

	return t
}

func procVersion(dm *discordgo.MessageCreate) error {
	cutoff := semver.MustParse("0.0.0")

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
			fmt.Printf("cutoff: %s\nversio: %s\n\n", cutoff, p.Latest.Time)
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
