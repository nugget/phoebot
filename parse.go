package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/nugget/phoebot/models"

	"github.com/blang/semver"
	"github.com/bwmarrin/discordgo"
)

type hookFunction func(*discordgo.MessageCreate) error

type Trigger struct {
	Regexp *regexp.Regexp
	Hook   hookFunction
	Direct bool
}

func LoadTriggers() error {
	triggers = append(triggers, regSubscriptions())
	triggers = append(triggers, regVersion())
	triggers = append(triggers, regTimezones())

	return nil
}

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

func regTimezones() (t Trigger) {
	exp := "(?i)((20[0-9][0-9]-[0-1][0-9]-[0-3][0-9] [0-2]?[0-9]:[0-5][0-9]) ([A-Z]+))"
	t.Regexp = regexp.MustCompile(exp)
	t.Hook = procTimezones
	t.Direct = false

	return t
}

func prettyTimezone(tz string) string {
	parts := strings.Split(tz, "/")
	tz = strings.ReplaceAll(parts[1], "_", " ")
	return tz
}

func smartLoc(tz string) (loc *time.Location) {
	var name string

	tz = strings.ToUpper(tz)

	switch tz {
	case "EDT", "EST":
		name = "America/New_York"
	case "CDT", "CST":
		name = "America/Chicago"
	case "MDT", "MST":
		name = "America/Denver"
	case "PDT", "PST":
		name = "America/Los_Angeles"
	case "BST", "GMT":
		name = "Europe/London"
	case "CET", "CEST":
		name = "Europe/Paris"
	case "AEST", "AEDT", "BRISBANE":
		name = "Australia/Brisbane"
	case "SYDNEY":
		name = "Australia/Sydney"
	case "ACDT", "ACST":
		name = "Australia/Adelaide"
	case "UTC":
		name = "UTC"
	default:
		name = "Unknown"
	}

	loc, err := time.LoadLocation(name)
	if err != nil {
		log.Printf("LoadLocation error: %v", err)
	}
	return loc
}

func procTimezones(dm *discordgo.MessageCreate) error {
	tzList := []string{
		"America/Los_Angeles",
		"America/New_York",
		"Europe/London",
		"Asia/Tokyo",
		"Australia/Adelaide",
		"Australia/Brisbane",
	}

	t := regTimezones()
	res := t.Regexp.FindStringSubmatch(dm.Content)
	//Dumper(res)

	parseLoc := smartLoc(res[3])
	log.Printf("parseLoc: %v", parseLoc)
	r, err := time.ParseInLocation("2006-01-02 15:04", res[2], parseLoc)
	if err != nil {
		log.Printf("time parse error: %v", err)
		return err
	}

	log.Printf("%v", r)

	mS := discordgo.MessageSend{}
	mE := discordgo.MessageEmbed{}

	mF := discordgo.MessageEmbedFooter{}
	mF.Text = r.Format(fmt.Sprintf("Converted from Mon 2-Jan-2006 15:04 MST (%s)", prettyTimezone(fmt.Sprintf("%s", parseLoc))))
	mE.Footer = &mF

	mE.Fields = make([]*discordgo.MessageEmbedField, 0)

	for _, l := range tzList {
		loc, err := time.LoadLocation(l)
		if err != nil {
			log.Printf("LoadLocation error: %v", err)
		}

		locName := fmt.Sprintf("%s", loc)
		timezone := r.In(loc).Format("MST")
		formatted := r.In(loc).Format("Mon 3:04PM")

		timezone = fmt.Sprintf("%s (%s)", prettyTimezone(locName), timezone)

		mE.Fields = append(mE.Fields, &discordgo.MessageEmbedField{
			Name:   timezone,
			Value:  formatted,
			Inline: true,
		})
		log.Printf("%s %s (%s)", formatted, timezone, loc)
	}

	mS.Embed = &mE

	s.Dg.ChannelMessageSendComplex(dm.ChannelID, &mS)

	return nil
}
