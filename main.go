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
	"time"

	"github.com/nugget/phoebot/hooks"
	"github.com/nugget/phoebot/lib/builddata"
	"github.com/nugget/phoebot/lib/config"
	"github.com/nugget/phoebot/lib/console"
	"github.com/nugget/phoebot/lib/coreprotect"
	"github.com/nugget/phoebot/lib/db"
	"github.com/nugget/phoebot/lib/discord"
	"github.com/nugget/phoebot/lib/ipc"
	"github.com/nugget/phoebot/lib/mcserver"
	"github.com/nugget/phoebot/lib/phoelib"
	"github.com/nugget/phoebot/lib/player"
	"github.com/nugget/phoebot/lib/postal"
	"github.com/nugget/phoebot/lib/products"
	"github.com/nugget/phoebot/lib/subscriptions"
	"github.com/nugget/phoebot/models"
	"github.com/nugget/phoebot/products/mojang"
	"github.com/nugget/phoebot/products/papermc"

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

func setupvc() *viper.Viper {
	c := viper.New()
	c.AutomaticEnv()
	c.SetDefault("MC_CHECK_INTERVAL", 600)
	c.SetDefault("MAILBOX_SCAN_INTERVAL", 7200)
	c.SetDefault("MAILBOX_POLL_INTERVAL", 300)
	c.SetDefault("MERCHANT_POLL_INTERVAL", 300)

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
		err := discord.SetPMOwner(dm)
		if err != nil {
			logrus.WithError(err).Error("Unable to update PM channel owner")
		}
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
		if t.InGame == true {
			// This is not a discord trigger
			continue
		}

		if direct == t.Direct || t.Direct == false {
			if t.Regexp.MatchString(dm.Content) {
				if t.ACL != "" && !phoelib.PlayerHasACL(dm.Author.ID, t.ACL) {
					logrus.WithFields(logrus.Fields{
						"player":   dm.Author.Username,
						"playerID": dm.Author.ID,
						"channel":  channel.Name,
						"direct":   direct,
						"t.direct": t.Direct,
						"regexp":   t.Regexp,
						"ACL":      t.ACL,
					}).Warn("Unauthorized Trigger")
				} else {
					logrus.WithFields(logrus.Fields{
						"player":   dm.Author.Username,
						"playerID": dm.Author.ID,
						"channel":  channel.Name,
						"direct":   direct,
						"t.direct": t.Direct,
						"regexp":   t.Regexp,
						"ACL":      t.ACL,
					}).Trace("Trigger matched, running hook")

					err := t.Hook(dm)
					if err != nil {
						logrus.WithFields(logrus.Fields{
							"player":   dm.Author.Username,
							"playerID": dm.Author.ID,
							"channel":  channel.Name,
							"direct":   direct,
							"t.direct": t.Direct,
							"regexp":   t.Regexp,
							"ACL":      t.ACL,
							"error":    err,
						}).Error("Error Hooking")
					}
				}
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
	triggers = append(triggers, hooks.RegCustomName())

	triggers = append(triggers, hooks.RegSubscriptions())
	triggers = append(triggers, hooks.RegUnsubAll())
	triggers = append(triggers, hooks.RegListSubscriptions())
	triggers = append(triggers, hooks.RegVersion())
	triggers = append(triggers, hooks.RegTimezones())
	triggers = append(triggers, hooks.RegStatus())
	triggers = append(triggers, hooks.RegServerInfo())
	triggers = append(triggers, hooks.RegServerList())

	triggers = append(triggers, hooks.RegSay())

	triggers = append(triggers, hooks.RegMapMe())
	triggers = append(triggers, hooks.RegNewMap())
	triggers = append(triggers, hooks.RegPOI())

	triggers = append(triggers, hooks.RegNearestPortal())

	triggers = append(triggers, hooks.RegLinkRequest())
	triggers = append(triggers, hooks.RegLinkVerify())

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
}

func SignLoop(interval int) error {
	for {
		playerCount, err := config.GetInt("players", 0)

		if playerCount == 0 {
			logrus.WithFields(logrus.Fields{
				"players": playerCount,
			}).Trace("Skipping SignLoop")

			time.Sleep(time.Duration(interval*2) * time.Second)

			continue
		}
		// Look for new tagging signs
		err = postal.NewSignScan()
		if err != nil {
			logrus.WithError(err).Error("postal.NewSignScan failed")
		}

		// Look for mailbox updates
		err = postal.PollMailboxes()
		if err != nil {
			logrus.WithError(err).Error("postal.PollMailboxes failed")
		}

		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func processWhisperStream(s *mcserver.Server) {
	for {
		w, stillOpen := <-ipc.ServerWhisperStream

		logrus.WithFields(logrus.Fields{
			"whisper":   w,
			"stillOpen": stillOpen,
		}).Info("ServerWhisperStream message received")

		err := s.Whisper(w.Who, w.Message)
		if err != nil {
			logrus.WithError(err).Error("processWhisperStream Error")
		}

		if !stillOpen {
			return
		}
	}
}

func processSayStream(s *mcserver.Server) {
	for {
		command, stillOpen := <-ipc.ServerSayStream

		logrus.WithFields(logrus.Fields{
			"command":   command,
			"stillOpen": stillOpen,
		}).Info("ServerSayStream message received")

		err := s.Say(command)
		if err != nil {
			logrus.WithError(err).Error("processSayStream Error")
		}

		if !stillOpen {
			return
		}
	}
}

func processChatStream(s *mcserver.Server) {
	for {
		c, stillOpen := <-ipc.ServerChatStream

		logrus.WithFields(logrus.Fields{
			"message":   c,
			"stillOpen": stillOpen,
		}).Trace("serverChatStream message received")

		noColorsMessage := c.ClearString()

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

		f := s.LogFields(logrus.Fields{
			"event":     "processChatStream",
			"translate": c.Translate,
			"class":     mcserver.ChatMsgClass(c),
		})

		// Process in-game triggers
		//
		triggerHits := 0

		for _, t := range triggers {
			if t.InGame == true {
				if t.Regexp.MatchString(noColorsMessage) {
					logrus.WithFields(logrus.Fields{
						"event":     "inGameTriggerHit",
						"translate": c.Translate,
						"class":     mcserver.ChatMsgClass(c),
						"message":   noColorsMessage,
						"regexp":    t.Regexp,
					}).Warn("InGame Trigger Match!")

					response, err := t.GameHook(noColorsMessage)
					if err != nil {
						logrus.WithFields(logrus.Fields{
							"error": err,
						}).Error("Error Hooking")
					}
					if response != "" {
						who, err := mcserver.GetPlayerNameFromWhisper(noColorsMessage)
						if err != nil {
							logrus.WithFields(logrus.Fields{
								"err": err,
								"who": who,
							}).Error("Unable to GetPlayerNameFromWhisper")
						} else {
							err := s.Whisper(who, response)
							if err != nil {
								logrus.WithFields(logrus.Fields{
									"err": err,
									"who": who,
								}).Error("Whisper failed")
							}
						}
					}
					triggerHits++
				}
			}
		}
		if triggerHits > 0 {
			// We hit triggers so no default action is desired with this
			// chatStream message
			logrus.Debug("No more chatStream processing since we hit at least one trigger.")
		} else {

			var (
				matchingSubs []models.Subscription
				err          error
				style        string
			)

			switch mcserver.ChatMsgClass(c) {
			case "whisper":
				logrus.WithFields(f).Info(noColorsMessage)
				matchingSubs, err = subscriptions.GetMatching("mcserver", "whispers")
			case "chat":
				logrus.WithFields(f).Debug(noColorsMessage)
				matchingSubs, err = subscriptions.GetMatching("mcserver", "chats")
			case "death":
				logrus.WithFields(f).Info(noColorsMessage)
				matchingSubs, err = subscriptions.GetMatching("mcserver", "deaths")
				style = "**"
			case "join":
				go StatsUpdate(s)
				logrus.WithFields(f).Info(noColorsMessage)
				matchingSubs, err = subscriptions.GetMatching("mcserver", "joins")
			case "announcement":
				logrus.WithFields(f).Warn(noColorsMessage)
			case "ignore":
				logrus.WithFields(f).Info(noColorsMessage)
			default:
				logrus.WithFields(f).Info(noColorsMessage)
				matchingSubs, err = subscriptions.GetMatching("mcserver", "events")
			}
			if err != nil {
				logrus.WithError(err).Warn("GetMatching failed on mcserver chat")
			} else {
				for _, sub := range matchingSubs {
					message := noColorsMessage

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
		}

		if !stillOpen {
			return
		}
	}
}

func StatsUpdate(s *mcserver.Server) error {
	ps, err := s.Status()
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
	vc := setupvc()
	builddata.LogConversational()

	for _, f := range []string{"DISCORD_BOT_TOKEN", "DATABASE_URI"} {
		tV := vc.GetString(f)
		if tV == "" {
			logrus.WithField("variable", f).Fatal("Missing environment variable")
		}
	}

	interval := vc.GetInt("MC_CHECK_INTERVAL")
	discordBotToken := vc.GetString("DISCORD_BOT_TOKEN")
	debugLevel := vc.GetString("PHOEBOT_DEBUG")
	dbURI := vc.GetString("DATABASE_URI")
	cpURI := vc.GetString("COREPROTECT_URI")

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

	err = coreprotect.Connect(cpURI)
	if err != nil {
		logrus.WithError(err).Fatal("Unable to connect to CoreProtect")
	}

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

	go products.Poller("PaperMC", "paper", interval, papermc.LatestVersion)
	go mojang.Poller(interval)

	go housekeeping(600)

	mc, err := mcserver.New()
	if err != nil {
		logrus.WithError(err).Fatal("Fatal mcserver error")
	}

	err = mc.Authenticate(
		vc.GetString("MINECRAFT_SERVER"),
		vc.GetInt("MINECRAFT_PORT"),
		vc.GetString("MOJANG_EMAIL"),
		vc.GetString("MOJANG_PASSWORD"),
	)
	if err != nil {
		logrus.WithError(err).Error("Error with mcserver Connect")
	}

	err = console.Initialize(
		vc.GetString("RCON_HOSTNAME"),
		vc.GetInt("RCON_PORT"),
		vc.GetString("RCON_PASSWORD"),
	)
	if err != nil {
		logrus.WithError(err).Error("RCON Initialization Failure")
	}

	{
		p, err := console.GetPlayer("MacNugget")
		logrus.WithFields(logrus.Fields{
			"player": p,
			"err":    err,
		}).Info("GetPlayer")
	}

	{
		s, err := console.GetServerInfo()
		logrus.WithFields(logrus.Fields{
			"server": s,
			"err":    err,
		}).Info("GetServerInfo")

	}

	go processChatStream(&mc)
	go processWhisperStream(&mc)
	go processSayStream(&mc)
	go StatsUpdate(&mc)
	go mc.Handler()

	// postal.Reset()

	mc.WaitForServer(5)

	go SignLoop(5)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	discord.Session.Close()

	shutdown()
}
