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
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/blang/semver"
	"github.com/nugget/phoebot/models"
	"github.com/nugget/phoebot/papermc"
	"github.com/nugget/phoebot/serverpro"
	"github.com/spf13/viper"

	"github.com/bwmarrin/discordgo"
)

type Subscription struct {
	ChannelID string  `xml:"channelID"`
	Product   Product `xml:"product"`
	Target    string  `xml:"target"`
}

type SubChannel struct {
	Operation string
	Sub       Subscription
}

type Check struct {
	Version semver.Version
	Time    time.Time
}

type State struct {
	LatestVersion map[string]Check `xml:"latestVersions"`
	Subscriptions []Subscription   `xml:"subscription"`
}

type DiscordMessage struct {
	ChannelID string
	Message   string
}

type Announcement struct {
	Product Product
	Message string
}

type Product struct {
	Name     string                       `xml:"name"`
	Class    string                       `xml:"class"`
	Type     string                       `xml:"type"`
	Function models.LatestVersionFunction `xml:"function"`
}

var (
	STATEFILE      string
	ProductList    []Product
	msgStream      chan DiscordMessage
	subStream      chan SubChannel
	announceStream chan Announcement
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

func SubscriptionsMatch(a, b Subscription) bool {
	if a.ChannelID == b.ChannelID {
		if a.Product.Name == b.Product.Name {
			return true
		}
	}

	return false
}

func (s *State) SubscriptionExists(sub Subscription) bool {
	for _, v := range s.Subscriptions {
		if SubscriptionsMatch(sub, v) {
			return true
		}
	}
	return false
}

func (s *State) AddSubscription(sub Subscription) error {
	if !s.SubscriptionExists(sub) {
		log.Printf("Adding %+v to subscription list", sub)
		s.Subscriptions = append(s.Subscriptions, sub)
		message := fmt.Sprintf("You are now subscribed to receive updates to this channel for %s releases on server.pro", sub.Product.Name)
		dg.ChannelMessageSend(sub.ChannelID, message)
	}

	//if sub.Product.Name == "Paper" {
	//	message := fmt.Sprintf("The highest version of PaperMC currently offered on server.pro is %s", s.PaperVersion)
	//	dg.ChannelMessageSend(sub.ChannelID, message)
	//}

	log.Printf("%d subscriptions in current state", len(s.Subscriptions))

	return nil
}

func (s *State) DropSubscription(sub Subscription) error {
	var newSubs []Subscription

	log.Printf("Dropping %+v from subscription list", sub)

	for _, v := range s.Subscriptions {
		if v.ChannelID != sub.ChannelID && v.Product.Name != sub.Product.Name {
			newSubs = append(newSubs, v)
		}
	}

	if len(s.Subscriptions) != len(newSubs) {
		message := fmt.Sprintf("You are no longer subscribed to receive updates to this channel for %s releases on server.pro", sub.Product.Name)
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
	// c.SetDefault("PRODUCTLIST", "Paper, Vanilla, CraftBukkit, Spigot, Vanilla Snapshot, Forge")

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
			if sub.Product.Name == d.Product.Name {
				if sub.Target != "" {
					d.Message = fmt.Sprintf("%s: %s", sub.Target, d.Message)
				}
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

func LookupProduct(serverType string) Product {
	for _, p := range ProductList {
		return p
	}

	return Product{}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	forMe := false
	for _, u := range m.Mentions {
		if u.ID == s.State.User.ID {
			forMe = true
		}
	}

	if forMe {
		subEx := regexp.MustCompile("(?i) ((un)?(sub)(scribe)?) ([^ ]+) ?(.*)")

		if subEx.MatchString(m.Content) {
			res := subEx.FindStringSubmatch(m.Content)
			log.Printf("(%d) %s", len(res), strings.Join(res, ":"))
			for i, v := range res {
				log.Printf("  %d: '%s'", i, v)
			}
			if len(res) == 7 {
				xUN := strings.ToLower(res[2])
				xSUB := strings.ToLower(res[3])
				serverType := strings.Title(res[5])
				target := res[6]

				product := LookupProduct(serverType)

				if xUN == "un" {
					sub := Subscription{m.ChannelID, product, target}
					subStream <- SubChannel{"DROP", sub}
				} else if xSUB == "sub" {
					sub := Subscription{m.ChannelID, product, target}
					subStream <- SubChannel{"ADD", sub}
				}
			}
		} else {
			message := fmt.Sprintf("I don't know what you're saying.  Try asking something like `subscribe paper [optional target]` or `unsubscribe paper`")
			s.ChannelMessageSend(m.ChannelID, message)
		}
	}
}

func (s *State) Looper(stream chan Announcement, product Product, interval int, fn models.LatestVersionFunction) {
	lastCheck := s.LatestVersion[product.Name]

	log.Printf("serverpro waiting for %s version > %s", product.Name, lastCheck.Version)

	for {
		maxVer, err := fn(product.Type)
		if err != nil {
			log.Printf("Error fetching %s Latest Version: %v", product.Name, err)
		} else {
			if maxVer.GT(lastCheck.Version) {
				message := fmt.Sprintf("Version %v of %v is available now", maxVer, product.Name)
				stream <- Announcement{product, message}
			} else {
				message := fmt.Sprintf("Version %v of %v is still the best", maxVer, product.Name)
				//stream <- Announcement{serverType, message}
				log.Printf(message)
			}

			s.LatestVersion[product.Name] = Check{maxVer, time.Now()}

		}

		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func LoadProducts(regFunc models.RegisterFunction, typesFunc models.GetTypesFunction) {
	class, fn := regFunc()
	typeList, err := typesFunc()
	if err != nil {
		log.Printf("Error fetching serverpro product list: %v", err)
	} else {
		for _, t := range typeList {
			p := Product{}
			p.Name = t
			p.Class = class
			p.Type = t
			p.Function = fn

			ProductList = append(ProductList, p)
		}
	}

	log.Printf("Loaded %d products to ProductList", len(ProductList))
	log.Printf("%+v", ProductList)
}

func main() {
	config := setupConfig()
	STATEFILE = config.GetString("STATE_FILENAME")

	currentState := State{}

	LoadProducts(serverpro.Register, serverpro.GetTypes)
	LoadProducts(papermc.Register, papermc.GetTypes)

	err := currentState.LoadState(STATEFILE)
	if err != nil {
		log.Printf("Unable to read state file: %v", err)
	}

	log.Printf("Loaded State: %+v", currentState)

	dg, err = discordgo.New("Bot " + config.GetString("DISCORD_BOT_TOKEN"))
	if err != nil {
		log.Fatalf("Error creating Discord session: ", err)
	}

	// fmt.Printf("\n%+v\n", dg)

	dg.AddHandler(messageCreate)

	err = dg.Open()
	if err != nil {
		log.Fatalf("Error opening Discord connection: ", err)
	}

	log.Printf("Connected to Discord as %s (SessionID %s)", dg.State.User, dg.State.SessionID)

	msgStream = make(chan DiscordMessage)
	subStream = make(chan SubChannel)
	announceStream = make(chan Announcement)

	go processMsgStream()
	go processSubStream(&currentState)
	go processAnnounceStream(&currentState)

	// go Looper(announceStream, serverpro.LatestVersion, "Paper", config.GetInt("MC_CHECK_INTERVAL"), &currentState)

	// msgStream <- DiscordMessage{"Moo", "Cow"}

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
