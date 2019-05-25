package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

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
	case "IST":
		name = "Asia/Kolkata"
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
		"Europe/Amsterdam",
		"Asia/Kolkata",
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
