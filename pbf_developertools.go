package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

func regLoglevel() (t Trigger) {
	t.Regexp = regexp.MustCompile("(?i)set loglevel to ([A-Z]+)")
	t.Hook = procLoglevel
	t.Direct = true

	return t
}

func procLoglevel(dm *discordgo.MessageCreate) error {
	t := regLoglevel()
	res := t.Regexp.FindStringSubmatch(dm.Content)

	Dumper(res)

	reqLevel := strings.ToLower(res[1])

	var setLevel logrus.Level

	switch reqLevel {
	case "trace":
		setLevel = logrus.TraceLevel
	case "debug":
		setLevel = logrus.DebugLevel
	case "info":
		setLevel = logrus.InfoLevel
	case "warn":
		setLevel = logrus.WarnLevel
	case "error":
		setLevel = logrus.ErrorLevel
	default:
		s.Dg.ChannelMessageSend(dm.ChannelID, "I don't recognize that log level.")
		return fmt.Errorf("Unrecognized loglevel '%s'", reqLevel)
	}

	oldLevel := logrus.GetLevel()
	logrus.SetLevel(setLevel)
	newLevel := logrus.GetLevel()

	if oldLevel != newLevel {
		logrus.WithFields(logrus.Fields{
			"old": oldLevel,
			"new": newLevel,
		}).Info("Console loging level changed")
	}

	message := fmt.Sprintf("Console logging level is '%s'", reqLevel)
	s.Dg.ChannelMessageSend(dm.ChannelID, message)

	return nil
}
