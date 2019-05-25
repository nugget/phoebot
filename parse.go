package main

import (
	"log"
	"regexp"
	"strings"

	"github.com/nugget/phoebot/models"
	"github.com/nugget/phoebot/state"

	"github.com/bwmarrin/discordgo"
)

type hookFunction func(state.State, *discordgo.MessageCreate) error

type Trigger struct {
	Regexp *regexp.Regexp
	Hook   hookFunction
	Direct bool
}

func LoadTriggers() error {
	triggers = append(triggers, regSubscriptions())
	triggers = append(triggers, regVersion())

	return nil
}

func regSubscriptions() (t Trigger) {
	t.Regexp = regexp.MustCompile("(?i)((un)?(sub)(scribe)?) ([^ ]+) ([^ ]+) ?(.*)")
	t.Hook = procSubscriptions
	t.Direct = true

	return t
}

func procSubscriptions(s state.State, dm *discordgo.MessageCreate) error {
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
			log.Printf("GetProduct error: %v", err)
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

func regVersion() (t Trigger) {
	t.Regexp = regexp.MustCompile("version report")
	t.Hook = procVersion
	t.Direct = true

	return t
}

func procVersion(s state.State, dm *discordgo.MessageCreate) error {
	mS := discordgo.MessageSend{}
	mS.Content = "Content"

	mE := discordgo.MessageEmbed{}
	mE.Description = "Description"

	mF := discordgo.MessageEmbedFooter{}
	mF.Text = "Footer Text"

	mE.Footer = &mF

	mE.Fields = make([]*discordgo.MessageEmbedField, 0)

	mE.Fields = append(mE.Fields, &discordgo.MessageEmbedField{
		Name:   "F1 Name",
		Value:  "F1 Value",
		Inline: true,
	})

	mE.Fields = append(mE.Fields, &discordgo.MessageEmbedField{
		Name:   "F2 Name",
		Value:  "F2 Value",
		Inline: true,
	})

	mS.Embed = &mE

	s.Dg.ChannelMessageSendComplex(dm.ChannelID, &mS)
	return nil
}
