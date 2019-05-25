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

	oldLevel := logrus.GetLevel()

	newLevel, err := LogLevel(reqLevel)
	if err != nil {
		logrus.WithError(err).Info("Failed to set LogLevel")
		s.Dg.ChannelMessageSend(dm.ChannelID, fmt.Sprintf("%s", err))
		return err
	}

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
