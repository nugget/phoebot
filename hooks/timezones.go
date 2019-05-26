package hooks

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/nugget/phoebot/lib/phoelib"
	"github.com/nugget/phoebot/lib/state"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

func RegTimezones() (t Trigger) {
	exp := "(?i)((20[0-9][0-9]-[0-1][0-9]-[0-3][0-9] [0-2]?[0-9]:[0-5][0-9]) ([A-Z]+))"
	t.Regexp = regexp.MustCompile(exp)
	t.Hook = ProcTimezones
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
		logrus.WithFields(logrus.Fields{
			"error": err,
			"name":  name,
			"tz":    tz,
		}).Error("Cannot load location")
	}
	return loc
}

func ProcTimezones(s *state.State, dm *discordgo.MessageCreate) error {
	tzList := []string{
		"America/Los_Angeles",
		"America/New_York",
		"Europe/London",
		"Europe/Amsterdam",
		"Asia/Kolkata",
		"Australia/Brisbane",
	}

	t := RegTimezones()
	res := t.Regexp.FindStringSubmatch(dm.Content)
	phoelib.DebugSlice(res)
	timeString := res[2] // This is the user-entered time string
	tzAbbr := res[3]     // This is the user-entered timezone abbreviation

	parseLoc := smartLoc(tzAbbr)

	logrus.WithFields(logrus.Fields{
		"tz":      tzAbbr,
		"pareLoc": parseLoc,
	}).Debug("SmartLoc mapped TZ to name")

	r, err := time.ParseInLocation("2006-01-02 15:04", timeString, parseLoc)
	if err != nil {
		return err
	}

	logrus.WithField("r", r).Debug("ParseInLocation successful")

	mS := discordgo.MessageSend{}
	mE := discordgo.MessageEmbed{}

	mF := discordgo.MessageEmbedFooter{}
	mF.Text = r.Format(fmt.Sprintf("Converted from Mon 2-Jan-2006 15:04 MST (%s)", prettyTimezone(fmt.Sprintf("%s", parseLoc))))
	mE.Footer = &mF

	mE.Fields = make([]*discordgo.MessageEmbedField, 0)

	for _, l := range tzList {
		loc, err := time.LoadLocation(l)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"l":     l,
				"error": err,
			}).Error("Unable to load timezone location")
		} else {
			locName := fmt.Sprintf("%s", loc)
			timezone := r.In(loc).Format("MST")
			formatted := r.In(loc).Format("Mon 3:04PM")

			timezone = fmt.Sprintf("%s (%s)", prettyTimezone(locName), timezone)

			mE.Fields = append(mE.Fields, &discordgo.MessageEmbedField{
				Name:   timezone,
				Value:  formatted,
				Inline: true,
			})
			logrus.WithFields(logrus.Fields{
				"locName":   locName,
				"timezone":  timezone,
				"formatted": formatted,
			}).Debug("Converted time")
		}
	}

	mS.Embed = &mE

	s.Dg.ChannelMessageSendComplex(dm.ChannelID, &mS)

	return nil
}
