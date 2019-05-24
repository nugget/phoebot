//
// Copyright (c) 2019 David McNett.  All Rights Reserved.
//
// SPDX-License-Identifier: BSD-2-Clause
//

package main

import (
	"encoding/xml"
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

type state struct {
	HostedVersion semver.Version `xml:"hostedVersion"`
	PaperVersion  semver.Version `xml:"paperVersion"`
}

func saveState(fileName string, s state) error {
	file, err := xml.MarshalIndent(s, "", " ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(fileName, file, 0644)
	return err
}

func loadState(fileName string) (s state, err error) {
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		return s, err
	}
	err = xml.Unmarshal(file, &s)

	return s, err
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

func main() {
	config := setupConfig()

	STATEFILE := config.GetString("STATE_FILENAME")

	currentState, err := loadState(STATEFILE)
	if err != nil {
		log.Printf("Unable to read state file: %v", err)
	}

	log.Printf("Loaded State: %+v", currentState)

	dg, err := discordgo.New("Bot " + config.GetString("DISCORD_BOT_TOKEN"))
	if err != nil {
		log.Fatalf("Error creating Discord session: ", err)
	}

	dg.AddHandler(messageCreate)

	err = dg.Open()
	if err != nil {
		log.Fatalf("Error opening Discord connection: ", err)
	}

	log.Printf("Connected to Discord as %s (SessionID %s)", dg.State.User, dg.State.SessionID)

	go serverpro.LoopLatestVersion("Paper", config.GetInt("MC_CHECK_INTERVAL"), &currentState.PaperVersion)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	dg.Close()

	err = saveState(STATEFILE, currentState)
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

	log.Printf("<%s> %s", m.Author, m.Content)
}
