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
	"regexp"
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
	msgStream      chan models.DiscordMessage
	subStream      chan models.SubChannel
	announceStream chan models.Announcement
)

func shutdown() {
	log.Printf("Shutting down...")
}

func setupConfig() *viper.Viper {
	c := viper.New()
	c.AutomaticEnv()
	c.SetDefault("MC_CHECK_INTERVAL", 600)
	c.SetDefault("STATE_FILENAME", "/phoebot/state.xml")
	// c.SetDefault("PRODUCTLIST", "Paper, Vanilla, CraftBukkit, Spigot, Vanilla Snapshot, Forge")

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
			if sub.Product.Name == d.Product.Name {
				if sub.Target != "" {
					d.Message = fmt.Sprintf("%s: %s", sub.Target, d.Message)
				}
				s.Dg.ChannelMessageSend(sub.ChannelID, d.Message)
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

	forMe := false
	for _, u := range dm.Mentions {
		if u.ID == ds.State.User.ID {
			forMe = true
		}
	}

	if forMe {
		subEx := regexp.MustCompile("(?i) ((un)?(sub)(scribe)?) ([^ ]+) ([^ ]+) ?(.*)")

		if subEx.MatchString(dm.Content) {
			res := subEx.FindStringSubmatch(dm.Content)
			Dumper(res)
			if len(res) == 8 {
				var err error

				sc := models.SubChannel{}

				xUN := strings.ToLower(res[2])
				xSUB := strings.ToLower(res[3])

				class := strings.ToLower(res[5])
				name := strings.Title(res[6])
				sc.Sub.Product, err = s.GetProduct(class, name)
				if err != nil {
					log.Printf("GetProduct error: %v", err)
					ds.ChannelMessageSend(dm.ChannelID, "I've never heard of that")
				} else {
					sc.Sub.ChannelID = dm.ChannelID
					sc.Sub.Target = res[7]

					if xUN == "un" {
						sc.Operation = "DROP"
					} else if xSUB == "sub" {
						sc.Operation = "ADD"
					}

					subStream <- sc
				}
			}
		} else {
			message := fmt.Sprintf("I don't know what you're saying.  Try asking something like `subscribe paper [optional target]` or `unsubscribe paper`")
			ds.ChannelMessageSend(dm.ChannelID, message)
		}

	}
}

func main() {
	config := setupConfig()
	STATEFILE = config.GetString("STATE_FILENAME")

	err := s.LoadState(STATEFILE)
	if err != nil {
		log.Printf("Unable to read state file: %v", err)
	}

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

	s.Dg, err = discordgo.New("Bot " + config.GetString("DISCORD_BOT_TOKEN"))
	if err != nil {
		log.Fatalf("Error creating Discord session: ", err)
	}

	// fmt.Printf("\n%+v\n\n", s.Products)

	s.Dg.AddHandler(messageCreate)

	err = s.Dg.Open()
	if err != nil {
		log.Fatalf("Error opening Discord connection: ", err)
	}

	log.Printf("Connected to Discord as %s (SessionID %s)", s.Dg.State.User, s.Dg.State.SessionID)

	msgStream = make(chan models.DiscordMessage)
	subStream = make(chan models.SubChannel)
	announceStream = make(chan models.Announcement)

	go processMsgStream()
	go processSubStream(&s)
	go processAnnounceStream(&s)

	for _, s := range s.Subscriptions {
		log.Printf("sub: %+v", s)
		// go s.Looper(announceStream, "serverpro", "Paper", 60, serverpro.LatestVersion)
	}

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
