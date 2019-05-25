package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/nugget/phoebot/models"
	"github.com/nugget/phoebot/state"

	"github.com/bwmarrin/discordgo"
)

type hookFunction func(*state.State, *discordgo.MessageCreate) error

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

func procSubscriptions(s *state.State, dm *discordgo.MessageCreate) error {
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

func procVersion(s *state.State, dm *discordgo.MessageCreate) error {
	cutoff := time.Now().Add(time.Duration(-1) * time.Hour)

	mS := discordgo.MessageSend{}

	mE := discordgo.MessageEmbed{}
	mE.Description = "Current Minecraft Versions:"

	mE.Fields = make([]*discordgo.MessageEmbedField, 0)

	for _, p := range s.Products {
		if p.Latest.Time.After(cutoff) {
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
