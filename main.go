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
	"syscall"

	"github.com/nugget/phoebot/hooks"
	"github.com/nugget/phoebot/lib/builddata"
	"github.com/nugget/phoebot/lib/db"
	"github.com/nugget/phoebot/lib/discord"
	"github.com/nugget/phoebot/lib/ipc"
	"github.com/nugget/phoebot/lib/phoelib"
	"github.com/nugget/phoebot/lib/player"
	"github.com/nugget/phoebot/lib/products"
	"github.com/nugget/phoebot/lib/subscriptions"
	"github.com/nugget/phoebot/products/papermc"
	"github.com/nugget/phoebot/products/serverpro"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	STATEFILE string
	triggers  []hooks.Trigger
)

func shutdown() {
	logrus.Infof("Shutting down...")
	os.Exit(0)
}

func setupConfig() *viper.Viper {
	c := viper.New()
	c.AutomaticEnv()
	c.SetDefault("MC_CHECK_INTERVAL", 600)

	return c
}

func processSubStream() {
	for {
		d, stillOpen := <-ipc.SubStream

		logrus.WithFields(logrus.Fields{
			"data":      d,
			"stillOpen": stillOpen,
		}).Debug("SubStream message received")

		if d.Operation == "DROP" {
			err := subscriptions.DropSubscription(d.Sub)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err,
					"sub":   d.Sub,
				}).Error("Unable to drop subscription")

				discord.Session.ChannelMessageSend(d.Sub.ChannelID, fmt.Sprintf("%v", err))
			}
		} else {
			err := subscriptions.AddSubscription(d.Sub)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err,
					"sub":   d.Sub,
				}).Error("Unable to add subscription")

				discord.Session.ChannelMessageSend(d.Sub.ChannelID, fmt.Sprintf("%v", err))
			}
		}

		if !stillOpen {
			return
		}
	}
}

func processAnnounceStream() {
	for {
		p, stillOpen := <-ipc.AnnounceStream

		logrus.WithFields(logrus.Fields{
			"p":         p,
			"stillOpen": stillOpen,
		}).Debug("announceStream message received")

		matchingSubs, err := subscriptions.GetMatching(p.Class, p.Name)
		if err != nil {
			logrus.WithError(err).Error("Unable to find matching subscriptions")
			return
		}

		logrus.WithFields(logrus.Fields{
			"class":   p.Class,
			"name":    p.Name,
			"version": p.Latest.Version,
		}).Warn("Announcing new version!")

		for _, sub := range matchingSubs {
			message := fmt.Sprintf("Version %v of %s on %s is available now", p.Latest.Version, p.Name, p.Class)

			if sub.Target != "" {
				message = fmt.Sprintf("%s: %s", sub.Target, message)
			}
			discord.Session.ChannelMessageSend(sub.ChannelID, message)
		}
		if !stillOpen {
			return
		}
	}
}

func processMsgStream() {
	for {
		d, stillOpen := <-ipc.MsgStream

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
	discord.RecordLog(dm)

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if dm.Author.ID == ds.State.User.ID {
		return
	}

	channel, err := discord.GetChannel(dm.ChannelID)
	if err != nil {
		logrus.WithError(err).Error("Unable to load channel info")
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

	player.UpdateFromDiscord(dm.Author)

	logrus.WithFields(logrus.Fields{
		"direct":    direct,
		"channel":   channel.Name,
		"username":  dm.Author.Username,
		"channelID": dm.ChannelID,
	}).Debug(logMsg)

	logrus.WithFields(logrus.Fields{
		"count":   len(triggers),
		"content": dm.Content,
	}).Trace("Evaluating triggers")

	for _, t := range triggers {

		if direct == t.Direct || t.Direct == false {
			if t.Regexp.MatchString(dm.Content) {
				logrus.WithFields(logrus.Fields{
					"direct":   direct,
					"t.direct": t.Direct,
					"regexp":   t.Regexp,
				}).Trace("Trigger matched, running hook")

				t.Hook(dm)
			} else {
				logrus.WithFields(logrus.Fields{
					"direct":   direct,
					"t.direct": t.Direct,
					"regexp":   t.Regexp,
				}).Trace("Trigger did not match")
			}
		} else {
			logrus.WithFields(logrus.Fields{
				"direct":   direct,
				"t.direct": t.Direct,
				"regexp":   t.Regexp,
			}).Trace("Ignored direct trigger in public")
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
	builddata.LogConversational()

	for _, f := range []string{"DISCORD_BOT_TOKEN", "DATABASE_URI"} {
		tV := config.GetString(f)
		if tV == "" {
			logrus.WithField("variable", f).Fatal("Missing environment variable")
		}
	}

	STATEFILE = config.GetString("STATE_FILENAME")

	interval := config.GetInt("MC_CHECK_INTERVAL")
	discordBotToken := config.GetString("DISCORD_BOT_TOKEN")
	debugLevel := config.GetString("PHOEBOT_DEBUG")
	dbURI := config.GetString("DATABASE_URI")

	if debugLevel != "" {
		_, err := phoelib.LogLevel(debugLevel)
		if err != nil {
			logrus.WithError(err).Error("Unable to set LogLevel")
		}
	}

	err := db.Connect(dbURI)
	if err != nil {
		logrus.WithError(err).Fatal("Unable to connect to database")
	}

	LoadTriggers()
	ipc.InitStreams()

	discord.Session, err = discordgo.New("Bot " + discordBotToken)
	if err != nil {
		logrus.WithError(err).Fatal("Unable to eonnect to Discord")
	}

	discord.Session.AddHandler(messageCreate)

	err = discord.Session.Open()
	if err != nil {
		logrus.WithError(err).Fatal("Error connecting to Discord")
	}

	logrus.WithFields(logrus.Fields{
		"user":      discord.Session.State.User,
		"sessionID": discord.Session.State.SessionID,
	}).Info("Connected to Discord")

	go processMsgStream()
	go processSubStream()
	go processAnnounceStream()

	go serverpro.Poller(interval)
	go products.Poller("PaperMC", "paper", interval, papermc.LatestVersion)

	// ipc.MsgStream <- DiscordMessage{"Moo", "Cow"}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	discord.Session.Close()

	shutdown()
}
