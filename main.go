//
// Copyright (c) 2019 David McNett.  All Rights Reserved.
//
// SPDX-License-Identifier: BSD-2-Clause
//

package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/nugget/phoebot/models"
	"github.com/nugget/phoebot/papermc"
	"github.com/nugget/phoebot/serverpro"
	"github.com/nugget/phoebot/state"
	"github.com/spf13/viper"

	"github.com/bwmarrin/discordgo"
)

var (
	s              state.State
	STATEFILE      string
	DEBUG          bool
	msgStream      chan models.DiscordMessage
	subStream      chan models.SubChannel
	announceStream chan models.Announcement
	triggers       []Trigger
)

func shutdown() {
	log.Printf("Shutting down...")
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
		log.Printf("subStream: %+v (%+v)", d, stillOpen)

		if d.Operation == "DROP" {
			err := s.DropSubscription(d.Sub)
			if err != nil {
				log.Printf("Error adding subscription: %v", err)
			}
		} else {
			err := s.AddSubscription(d.Sub)
			if err != nil {
				log.Printf("Error adding subscription: %v", err)
			}
		}

		err := s.SaveState(STATEFILE)
		if err != nil {
			log.Printf("Error saving state: %v", err)
		}

		if !stillOpen {
			return
		}
	}
}

func processAnnounceStream(s *state.State) {
	for {
		d, stillOpen := <-announceStream
		log.Printf("announceStream: %+v (%+v)", d, stillOpen)

		for _, sub := range s.Subscriptions {
			if strings.ToLower(sub.Class) == strings.ToLower(d.Product.Class) {
				if strings.ToLower(sub.Name) == strings.ToLower(d.Product.Name) {
					if sub.Target != "" {
						d.Message = fmt.Sprintf("%s: %s", sub.Target, d.Message)
					}
					s.Dg.ChannelMessageSend(sub.ChannelID, d.Message)
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
		log.Printf("msgStream: %+v (%+v)", d, stillOpen)
		if !stillOpen {
			return
		}
	}
}

func Dumper(res []string) {
	log.Printf("(%d) %s", len(res), strings.Join(res, ":"))
	for i, v := range res {
		log.Printf("  %d: '%s'", i, v)
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
		log.Printf("Unable to load channel info: %v", err)
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

	origin := fmt.Sprintf("#%s <%s>", channel.Name, dm.Author.Username)

	log.Printf("(%d) [%v] >> %v %+v", len(triggers), direct, origin, dm.Content)

	for _, t := range triggers {
		if direct == t.Direct || t.Direct == false {
			if t.Regexp.MatchString(dm.Content) {
				t.Hook(dm)
			}
		}
	}
	log.Printf("-- ")
}

func main() {
	config := setupConfig()
	STATEFILE = config.GetString("STATE_FILENAME")
	INTERVAL := config.GetInt("MC_CHECK_INTERVAL")
	DEBUG = config.GetBool("PHOEBOT_DEBUG")

	err := s.LoadState(STATEFILE)
	if err != nil {
		log.Printf("Unable to read state file: %v", err)
	}

	LoadTriggers()

	s.Dg, err = discordgo.New("Bot " + config.GetString("DISCORD_BOT_TOKEN"))
	if err != nil {
		log.Fatalf("Error creating Discord session: ", err)
	}

	s.Dg.AddHandler(messageCreate)

	log.Printf("Loaded State from %s", STATEFILE)

	s.DedupeProducts()

	err = s.LoadProducts(serverpro.Register, serverpro.GetTypes)
	if err != nil {
		log.Printf("Error loading products: %v", err)
	}

	err = s.LoadProducts(papermc.Register, papermc.GetTypes)
	if err != nil {
		log.Printf("Error loading products: %v", err)
	}

	err = s.Dg.Open()
	if err != nil {
		log.Fatalf("Error opening Discord connection: %v", err)
	}

	log.Printf("Connected to Discord as %s (SessionID %s)", s.Dg.State.User, s.Dg.State.SessionID)

	msgStream = make(chan models.DiscordMessage)
	subStream = make(chan models.SubChannel)
	announceStream = make(chan models.Announcement)

	go processMsgStream()
	go processSubStream(&s)
	go processAnnounceStream(&s)

	go s.Looper(announceStream, "server.pro", "Paper", INTERVAL, serverpro.LatestVersion)
	go s.Looper(announceStream, "server.pro", "Vanilla", INTERVAL, serverpro.LatestVersion)
	go s.Looper(announceStream, "PaperMC", "paper", INTERVAL, papermc.LatestVersion)

	// msgStream <- DiscordMessage{"Moo", "Cow"}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	s.Dg.Close()

	err = s.SaveState(STATEFILE)
	if err != nil {
		log.Printf("Error writing state file: %v", err)
	} else {
		log.Printf("Saved state to %s", STATEFILE)
	}

	shutdown()
}
