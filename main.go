//
// Copyright (c) 2019 David McNett.  All Rights Reserved.
//
// SPDX-License-Identifier: BSD-2-Clause
//

package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/blang/semver"
	"github.com/nugget/phoebot/serverpro"
	"github.com/spf13/viper"

	"github.com/bwmarrin/discordgo"
)

type Subscription struct {
	ChannelID  string `xml:"channelID"`
	ServerType string `xml:"serverType"`
}

type SubChannel struct {
	Operation string
	Sub       Subscription
}

type State struct {
	HostedVersion semver.Version `xml:"hostedVersion"`
	PaperVersion  semver.Version `xml:"paperVersion"`
	Subscriptions []Subscription `xml:"subscription"`
}

type DiscordMessage struct {
	ChannelID string
	Message   string
}

var (
	STATEFILE      string
	msgStream      chan DiscordMessage
	subStream      chan SubChannel
	announceStream chan serverpro.Announcement
	dg             *discordgo.Session
)

func (s *State) SaveState(fileName string) error {
	file, err := xml.MarshalIndent(s, "", " ")
	if err != nil {
		return err
	}

	fmt.Printf("-- \n%s\n-- \n", string(file))

	err = ioutil.WriteFile(fileName, file, 0644)
	return err
}

func (s *State) LoadState(fileName string) (err error) {
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}
	err = xml.Unmarshal(file, &s)

	fmt.Printf("-- \n%s\n-- \n", string(file))

	return err
}

func (s *State) SubscriptionExists(sub Subscription) bool {
	for _, v := range s.Subscriptions {
		if v == sub {
			return true
		}
	}
	return false
}

func (s *State) AddSubscription(sub Subscription) error {
	if !s.SubscriptionExists(sub) {
		log.Printf("Adding %+v to subscription list", sub)
		s.Subscriptions = append(s.Subscriptions, sub)
		message := fmt.Sprintf("You are now subscribed to receive updates to this channel for %s releases on server.pro", sub.ServerType)
		dg.ChannelMessageSend(sub.ChannelID, message)
	}

	if sub.ServerType == "Paper" {
		message := fmt.Sprintf("The highest version of PaperMC currently offered on server.pro is %s", s.PaperVersion)
		dg.ChannelMessageSend(sub.ChannelID, message)
	}

	log.Printf("%d subscriptions in current state", len(s.Subscriptions))

	return nil
}

func (s *State) DropSubscription(sub Subscription) error {
	var newSubs []Subscription

	log.Printf("Dropping %+v from subscription list", sub)

	for _, v := range s.Subscriptions {
		if v != sub {
			newSubs = append(newSubs, v)
		}
	}

	if len(s.Subscriptions) != len(newSubs) {
		message := fmt.Sprintf("You are no longer subscribed to receive updates to this channel for %s releases on server.pro", sub.ServerType)
		dg.ChannelMessageSend(sub.ChannelID, message)
	}

	s.Subscriptions = newSubs

	log.Printf("%d subscriptions in current state", len(s.Subscriptions))

	return nil
}

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

func processSubStream(s *State) {
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

func processAnnounceStream(s *State) {
	for {
		d, stillOpen := <-announceStream
		log.Printf("announceStream: %+v (%+v)", d, stillOpen)

		for _, sub := range s.Subscriptions {
			if sub.ServerType == d.ServerType {
				dg.ChannelMessageSend(sub.ChannelID, d.Message)
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

func main() {
	config := setupConfig()
	STATEFILE = config.GetString("STATE_FILENAME")

	currentState := State{}

	err := currentState.LoadState(STATEFILE)
	if err != nil {
		log.Printf("Unable to read state file: %v", err)
	}

	log.Printf("Loaded State: %+v", currentState)

	dg, err = discordgo.New("Bot " + config.GetString("DISCORD_BOT_TOKEN"))
	if err != nil {
		log.Fatalf("Error creating Discord session: ", err)
	}

	dg.AddHandler(messageCreate)

	err = dg.Open()
	if err != nil {
		log.Fatalf("Error opening Discord connection: ", err)
	}

	log.Printf("Connected to Discord as %s (SessionID %s)", dg.State.User, dg.State.SessionID)

	msgStream = make(chan DiscordMessage)
	subStream = make(chan SubChannel)
	announceStream = make(chan serverpro.Announcement)

	go processMsgStream()
	go processSubStream(&currentState)
	go processAnnounceStream(&currentState)

	go serverpro.LoopLatestVersion(announceStream, "Paper", config.GetInt("MC_CHECK_INTERVAL"), &currentState.PaperVersion)

	msgStream <- DiscordMessage{"Moo", "Cow"}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	dg.Close()

	err = currentState.SaveState(STATEFILE)
	if err != nil {
		log.Printf("Error writing state file: %v", err)
	} else {
		log.Printf("Saved state to %s", STATEFILE)
	}

	shutdown()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}
	// If the message is "ping" reply with "Pong!"
	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	}

	// If the message is "pong" reply with "Ping!"
	if m.Content == "pong" {
		s.ChannelMessageSend(m.ChannelID, "Ping!")
	}

	if m.Content == "papersubscribe" {
		sub := Subscription{m.ChannelID, "Paper"}
		subStream <- SubChannel{"ADD", sub}
	}

	if m.Content == "paperunsubscribe" {
		sub := Subscription{m.ChannelID, "Paper"}
		subStream <- SubChannel{"DROP", sub}
	}

	log.Printf("<%s> %s", m.Author, m.Content)
}
