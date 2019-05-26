//
// Copyright (c) 2019 David McNett.  All Rights Reserved.
//
// SPDX-License-Identifier: BSD-2-Clause
//

package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/nugget/phoebot/hooks"
	"github.com/nugget/phoebot/lib/state"
	"github.com/nugget/phoebot/models"
	"github.com/nugget/phoebot/products/papermc"
	"github.com/nugget/phoebot/products/serverpro"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	PB             state.State
	STATEFILE      string
	msgStream      chan models.DiscordMessage
	subStream      chan models.SubChannel
	announceStream chan models.Announcement
	triggers       []models.Trigger
)

func shutdown() {
	logrus.Infof("Shutting down...")
	os.Exit(0)
}

func setupConfig() *viper.Viper {
	c := viper.New()
	c.AutomaticEnv()
	c.SetDefault("MC_CHECK_INTERVAL", 600)
	c.SetDefault("STATE_FILENAME", "/phoebot/state.xml")

	return c
}

func processSubStream(s *state.State) {
	for {
		d, stillOpen := <-subStream

		logrus.WithFields(logrus.Fields{
			"data":      d,
			"stillOpen": stillOpen,
		}).Debug("subStream message received")

		if d.Operation == "DROP" {
			err := PB.DropSubscription(d.Sub)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err,
					"sub":   d.Sub,
				}).Error("Unable to drop subscription")

				PB.Dg.ChannelMessageSend(d.Sub.ChannelID, fmt.Sprintf("%v", err))
			}
		} else {
			err := s.AddSubscription(d.Sub)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err,
					"sub":   d.Sub,
				}).Error("Unable to add subscription")

				PB.Dg.ChannelMessageSend(d.Sub.ChannelID, fmt.Sprintf("%v", err))
			}
		}

		err := PB.SaveState(STATEFILE)
		if err != nil {
			logrus.WithError(err).Error("Unable to write state file from subStream")
		}

		if !stillOpen {
			return
		}
	}
}

func processAnnounceStream(s *state.State) {
	for {
		d, stillOpen := <-announceStream

		logrus.WithFields(logrus.Fields{
			"data":      d,
			"stillOpen": stillOpen,
		}).Debug("announceStream message received")

		for _, sub := range s.Subscriptions {
			if strings.ToLower(sub.Class) == strings.ToLower(d.Product.Class) {
				if strings.ToLower(sub.Name) == strings.ToLower(d.Product.Name) {
					if sub.Target != "" {
						d.Message = fmt.Sprintf("%s: %s", sub.Target, d.Message)
					}
					PB.Dg.ChannelMessageSend(sub.ChannelID, d.Message)
				}
			}
		}
		if !stillOpen {
			return
		}
	}
}

func processMsgStream() {
	for {
		d, stillOpen := <-msgStream

		logrus.WithFields(logrus.Fields{
			"data":      d,
			"stillOpen": stillOpen,
		}).Debug("messageStream message received")

		if !stillOpen {
			return
		}
	}
}

func messageCreate(ds *discordgo.Session, dm *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if dm.Author.ID == ds.State.User.ID {
		return
	}

	channel, err := ds.State.Channel(dm.ChannelID)
	if err != nil {
		logrus.WithError(err).Error("Unavle to load channel info")
		channel.Name = "unknown"
	}

	direct := false

	for _, u := range dm.Mentions {
		if u.ID == ds.State.User.ID {
			direct = true
		}
	}

	if channel.Type == 1 && channel.Name == "" {
		// This is a private message window
		direct = true
		channel.Name = "PM"
	}

	logMsg := fmt.Sprintf("<%s> %s", dm.Author.Username, dm.Content)

	logrus.WithFields(logrus.Fields{
		"direct":    direct,
		"channel":   channel.Name,
		"username":  dm.Author.Username,
		"channelID": dm.ChannelID,
	}).Debug(logMsg)

	for _, t := range triggers {
		if direct == t.Direct || t.Direct == false {
			if t.Regexp.MatchString(dm.Content) {
				t.Hook(dm)
			}
		}
	}
}

type hookFunction func(*discordgo.MessageCreate) error

func LoadTriggers() error {
	// This is the baseline feature you can use to pattern new features you
	// want to add
	triggers = append(triggers, hooks.RegTemplate())
	triggers = append(triggers, hooks.RegLoglevel())

	triggers = append(triggers, hooks.RegSubscriptions())
	triggers = append(triggers, hooks.RegListSubscriptions())
	triggers = append(triggers, hooks.RegVersion())
	triggers = append(triggers, hooks.RegTimezones())
	triggers = append(triggers, hooks.RegStatus())

	return nil
}

func main() {
	config := setupConfig()

	STATEFILE = config.GetString("STATE_FILENAME")

	interval := config.GetInt("MC_CHECK_INTERVAL")
	discordBotToken := config.GetString("DISCORD_BOT_TOKEN")
	debugLevel := config.GetString("PHOEBOT_DEBUG")

	if debugLevel != "" {
		_, err := LogLevel(debugLevel)
		if err != nil {
			logrus.WithError(err).Error("Unable to set LogLevel")
		}
	}

	for _, f := range []string{"DISCORD_BOT_TOKEN"} {
		tV := config.GetString(f)
		if tV == "" {
			logrus.WithField("variable", f).Fatal("Missing environment variable")
		}
	}

	err := s.LoadState(STATEFILE)
	if err != nil {
		logrus.WithError(err).Warn("Unable to read state file")
	} else {
		logrus.WithField("filename", STATEFILE).Infof("Loaded saved state")
	}

	LoadTriggers()

	PB.Dg, err = discordgo.New("Bot " + discordBotToken)
	if err != nil {
		logrus.WithError(err).Fatal("Unable to eonnect to Discord")
	}

	PB.Dg.AddHandler(messageCreate)

	PB.DedupeProducts()

	err = PB.LoadProducts(serverpro.Register, serverpro.GetTypes)
	if err != nil {
		logrus.WithError(err).Warn("Error loading products from serverpro")
	}

	err = PB.LoadProducts(papermc.Register, papermc.GetTypes)
	if err != nil {
		logrus.WithError(err).Warn("Error loading products from papermc")
	}

	err = PB.Dg.Open()
	if err != nil {
		logrus.WithError(err).Fatal("Error connecting to Discord")
	}

	logrus.WithFields(logrus.Fields{
		"user":      PB.Dg.State.User,
		"sessionID": PB.Dg.State.SessionID,
	}).Info("Connected to Discord")

	msgStream = make(chan models.DiscordMessage)
	subStream = make(chan models.SubChannel)
	announceStream = make(chan models.Announcement)

	go processMsgStream()
	go processSubStream(&s)
	go processAnnounceStream(&s)

	go PB.ProductPoller(announceStream, "server.pro", "Paper", interval, serverpro.LatestVersion)
	go PB.ProductPoller(announceStream, "server.pro", "Vanilla", interval, serverpro.LatestVersion)
	go PB.ProductPoller(announceStream, "PaperMC", "paper", interval, papermc.LatestVersion)

	// msgStream <- DiscordMessage{"Moo", "Cow"}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	PB.Dg.Close()

	err = PB.SaveState(STATEFILE)
	if err != nil {
		logrus.WithError(err).Error("Unable to write state file")
	} else {
		logrus.WithField("filename", STATEFILE).Info("Saved state to disk")
	}

	shutdown()
}
