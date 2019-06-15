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
	"regexp"
	"syscall"
	"time"

	"github.com/nugget/phoebot/hooks"
	"github.com/nugget/phoebot/lib/builddata"
	"github.com/nugget/phoebot/lib/db"
	"github.com/nugget/phoebot/lib/discord"
	"github.com/nugget/phoebot/lib/ipc"
	"github.com/nugget/phoebot/lib/mcserver"
	"github.com/nugget/phoebot/lib/phoelib"
	"github.com/nugget/phoebot/lib/player"
	"github.com/nugget/phoebot/lib/products"
	"github.com/nugget/phoebot/lib/subscriptions"
	"github.com/nugget/phoebot/models"
	"github.com/nugget/phoebot/products/mojang"
	"github.com/nugget/phoebot/products/papermc"
	"github.com/nugget/phoebot/products/serverpro"

	"github.com/Tnze/go-mc/chat"
	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	triggers []hooks.Trigger
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
		}).Debug("AnnounceStream message received")

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

func processMojangStream() {
	for {
		a, stillOpen := <-ipc.MojangStream

		logrus.WithFields(logrus.Fields{
			"a":         a,
			"stillOpen": stillOpen,
		}).Debug("MojangStream message received")

		matchingSubs, err := subscriptions.GetMatching("mojang", a.Product)
		if err != nil {
			logrus.WithError(err).Error("Unable to find matching subscriptions")
			return
		}

		logrus.WithFields(logrus.Fields{
			"title":   a.Title,
			"date":    a.PublishDate,
			"product": a.Product,
			"Version": a.Version,
			"release": a.Release,
			"url":     a.URL,
		}).Warn("Announcing new mojang article!")

		for _, sub := range matchingSubs {
			message := fmt.Sprintf("New Minecraft %s announced at %s", a.Version, a.URL)

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

	if phoelib.IgnoreMessage(dm) {
		//logrus.WithField("logMsg", logMsg).Trace("Ignored Discord message")
		return
	}

	err = player.UpdateFromDiscord(dm.Author)
	if err != nil {
		logrus.WithError(err).Error("player.UpdateFromDiscord failed")
	}

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
	triggers = append(triggers, hooks.RegUnsubAll())
	triggers = append(triggers, hooks.RegListSubscriptions())
	triggers = append(triggers, hooks.RegVersion())
	triggers = append(triggers, hooks.RegTimezones())
	triggers = append(triggers, hooks.RegStatus())

	return nil
}

func housekeeping(interval int) error {
	for {
		err := phoelib.LoadIgnores()
		if err != nil {
			logrus.WithError(err).Error("LoadIgnores failed")
		}

		time.Sleep(time.Duration(interval) * time.Second)
	}

	return nil
}

func OnChatMsg(c chat.Message, pos byte) error {
	coloredMessage := c.String()
	cleanMessage := mcserver.CleanString(c)

	re := regexp.MustCompile(`\x1B\[[0-?]*[ -/]*[@-~]`)
	if re.MatchString(coloredMessage) {
		cleanMessage = re.ReplaceAllString(coloredMessage, "")
	}

	for i, e := range c.Extra {
		logrus.WithFields(logrus.Fields{
			"i":             i,
			"text":          e.Text,
			"bold":          e.Bold,
			"italic":        e.Italic,
			"underlined":    e.UnderLined,
			"strikethrough": e.StrikeThrough,
			"obfuscated":    e.Obfuscated,
			"color":         e.Color,
		}).Trace("onChatMsg Debug Extra")
	}

	for i, w := range c.With {
		logrus.WithFields(logrus.Fields{
			"i": i,
			"w": string(w),
		}).Trace("onChatMsg Debug With")
	}

	f := mcserver.LogFields(logrus.Fields{
		"pos":       pos,
		"event":     "OnChatMsg",
		"translate": c.Translate,
		"class":     mcserver.ChatMsgClass(c),
	})

	var (
		matchingSubs []models.Subscription
		err          error
		style        string
	)

	switch mcserver.ChatMsgClass(c) {
	case "whisper":
		logrus.WithFields(f).Info(cleanMessage)
		matchingSubs, err = subscriptions.GetMatching("mcserver", "whispers")
	case "chat":
		logrus.WithFields(f).Debug(cleanMessage)
		matchingSubs, err = subscriptions.GetMatching("mcserver", "chats")
	case "death":
		logrus.WithFields(f).Info(cleanMessage)
		matchingSubs, err = subscriptions.GetMatching("mcserver", "deaths")
		style = "**"
	case "join":
		go StatsUpdate()
		logrus.WithFields(f).Info(cleanMessage)
		matchingSubs, err = subscriptions.GetMatching("mcserver", "joins")
	case "announcement":
		logrus.WithFields(f).Warn(cleanMessage)
	case "ignore":
		logrus.WithFields(f).Warn(cleanMessage)
	default:
		logrus.WithFields(f).Info(cleanMessage)
		matchingSubs, err = subscriptions.GetMatching("mcserver", "events")
	}
	if err != nil {
		logrus.WithError(err).Warn("GetMatching failed on mcserver chat")
	} else {
		for _, sub := range matchingSubs {
			message := cleanMessage

			if sub.Target != "" {
				message = fmt.Sprintf("%s: %s", sub.Target, message)
			}
			logrus.WithFields(logrus.Fields{
				"channel": sub.ChannelID,
				"message": message,
			}).Debug("ChannelMessageSend")

			discord.Session.ChannelMessageSend(sub.ChannelID, style+message+style)
		}
	}

	return nil
}

func StatsUpdate() error {
	ps, err := mcserver.GetPingStats()
	if err != nil {
		logrus.WithError(err).Error("PingAndList Failure")
	} else {
		logrus.WithFields(logrus.Fields{
			"delay":      ps.Delay,
			"online":     ps.PlayersOnline,
			"max":        ps.PlayersMax,
			"version":    ps.Version,
			"serverName": ps.Description,
		}).Debug("PingAndList")

		newTopic := fmt.Sprintf("%s (%d/%d players online)", ps.Description, ps.PlayersOnline, ps.PlayersMax)

		matchingSubs, err := subscriptions.GetMatching("mcserver", "topic")
		if err != nil {
			logrus.WithError(err).Warn("GetMatching failed on mcserver chat")
		} else {
			for _, sub := range matchingSubs {
				// discord.Session.ChannelMessageSend(sub.ChannelID, newTopic)

				cE := discordgo.ChannelEdit{}
				cE.Topic = newTopic

				_, err := discord.Session.ChannelEditComplex(sub.ChannelID, &cE)
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"error":     err,
						"channelID": sub.ChannelID,
						"topic":     newTopic,
					}).Error("ChannelEditComplex Failure")
				} else {
					logrus.WithFields(logrus.Fields{
						"channelID": sub.ChannelID,
						"topic":     newTopic,
					}).Info("Set Channel Topic")
				}
			}

		}
	}

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
	go processMojangStream()

	go serverpro.Poller(interval)
	go products.Poller("PaperMC", "paper", interval, papermc.LatestVersion)
	go mojang.Poller(interval)

	go housekeeping(600)

	// ipc.MsgStream <- DiscordMessage{"Moo", "Cow"}
	//
	err = mcserver.Login(
		config.GetString("MINECRAFT_SERVER"),
		config.GetInt("MINECRAFT_PORT"),
		config.GetString("MOJANG_EMAIL"),
		config.GetString("MOJANG_PASSWORD"),
	)
	if err != nil {
		logrus.WithError(err).Error("Error connecting to Minecraft")
	} else {
		mcserver.Client.Events.GameStart = mcserver.OnGameStart
		mcserver.Client.Events.ChatMsg = OnChatMsg
		mcserver.Client.Events.Disconnect = mcserver.OnDisconnect
		mcserver.Client.Events.PluginMessage = mcserver.OnPluginMessage
		mcserver.Client.Events.Die = mcserver.OnDieMessage

		go mcserver.Handler()
		go StatsUpdate()
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	discord.Session.Close()

	shutdown()
}
