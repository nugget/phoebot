package main

import (
	logger "log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nugget/phoebot/serverpro"

	"github.com/bwmarrin/discordgo"
)

var (
	latestVersion string = "0.0.1"
)

func main() {
	dg, err := discordgo.New("Bot " + os.Getenv("DISCORD_BOT_TOKEN"))
	if err != nil {
		logger.Fatalf("Error creating Discord session: ", err)
	}

	dg.AddHandler(messageCreate)

	err = dg.Open()
	if err != nil {
		logger.Fatalf("Error opening Discord connection: ", err)
	}

	logger.Printf("Connected to Discord as %s (SessionID %s)", dg.State.User, dg.State.SessionID)

	lv, err := serverpro.LatestVersion("Paper")
	if err != nil {
		logger.Printf("Unable to get latest version: %v", err)
	} else {
		logger.Printf("lv: %+v", lv)
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	dg.Close()
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

	logger.Printf("<%s> %s", m.Author, m.Content)
}
