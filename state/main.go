package state

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/nugget/phoebot/models"

	"github.com/bwmarrin/discordgo"
)

type State struct {
	Products      []models.Product      `xml:"product"`
	Subscriptions []models.Subscription `xml:"subscription"`
	Dg            *discordgo.Session    `xml:"-"`
}

func (s *State) SaveState(fileName string) error {
	file, err := xml.MarshalIndent(s, "", " ")
	if err != nil {
		return err
	}

	//fmt.Printf("-- \n%s\n-- \n", string(file))

	err = ioutil.WriteFile(fileName, file, 0644)
	return err
}

func (s *State) LoadState(fileName string) (err error) {
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}
	err = xml.Unmarshal(file, &s)

	//fmt.Printf("-- \n%s\n-- \n", string(file))

	return err
}

func SubscriptionsMatch(a, b models.Subscription) bool {
	if a.ChannelID == b.ChannelID {
		if strings.ToLower(a.Class) == strings.ToLower(b.Class) {
			if strings.ToLower(a.Name) == strings.ToLower(b.Name) {
				return true
			}
		}
	}

	return false
}

func (s *State) SubscriptionExists(sub models.Subscription) bool {
	for _, v := range s.Subscriptions {
		if SubscriptionsMatch(sub, v) {
			return true
		}
	}
	return false
}

func (s *State) AddSubscription(sub models.Subscription) error {
	if sub.Class == "" || sub.Name == "" {
		return fmt.Errorf("Cannot add malformed subscription: %s/%s", sub.Class, sub.Name)
	} else {
		if !s.SubscriptionExists(sub) {
			log.Printf("Adding %+v to subscription list", sub)
			s.Subscriptions = append(s.Subscriptions, sub)
			message := fmt.Sprintf("You are now subscribed to receive updates to this channel for %s releases from %s", sub.Name, sub.Class)
			s.Dg.ChannelMessageSend(sub.ChannelID, message)
		}

		log.Printf("%d subscriptions in current state", len(s.Subscriptions))
	}

	return nil
}

func (s *State) DropSubscription(sub models.Subscription) error {
	var newSubs []models.Subscription

	log.Printf("Dropping %+v from subscription list", sub)

	for _, v := range s.Subscriptions {
		if !SubscriptionsMatch(sub, v) {
			newSubs = append(newSubs, v)
		}
	}

	if len(s.Subscriptions) != len(newSubs) {
		message := fmt.Sprintf("You are no longer subscribed to receive updates to this channel for %s releases from %s", sub.Name, sub.Class)
		s.Dg.ChannelMessageSend(sub.ChannelID, message)
	}

	s.Subscriptions = newSubs

	log.Printf("%d subscriptions in current state", len(s.Subscriptions))

	return nil
}

func (s *State) GetProduct(class, name string) (models.Product, error) {
	for _, p := range s.Products {
		if strings.ToLower(p.Class) == strings.ToLower(class) {
			if strings.ToLower(p.Name) == strings.ToLower(name) {
				return p, nil
			}
		}
	}

	return models.Product{}, fmt.Errorf("Product not found")
}

func (s *State) PutProduct(n models.Product) error {
	if n.Class == "" || n.Name == "" || n.Function == nil {
		return fmt.Errorf("Can't load malformed product: %+v", n)
	}

	newProducts := []models.Product{}

	added := false
	for _, p := range s.Products {

		if p.Class == "" || p.Name == "" {
			log.Printf("Not putting malformed product: %+v", p)
			// Skip this one
		} else if p.Class == n.Class && p.Name == n.Name {
			if n.Latest.Time.After(p.Latest.Time) {
				newProducts = append(newProducts, n)
			} else {
				newProducts = append(newProducts, p)
			}
			added = true
		} else {
			newProducts = append(newProducts, p)
		}
	}

	if !added {
		newProducts = append(newProducts, n)
	}

	s.Products = newProducts

	return nil
}

func (s *State) DedupeProducts() error {
	newProducts := []models.Product{}
	exists := make(map[string]bool)

	for _, p := range s.Products {
		key := fmt.Sprintf("%v-%v", p.Class, p.Name)
		if !exists[key] {
			if p.Class == "" || p.Name == "" {
				log.Printf("Skipping malformed product during dedupe: %+v", p)
			} else {
				newProducts = append(newProducts, p)
			}
		}
		exists[key] = true
	}

	if len(s.Products) != len(newProducts) {
		log.Printf("Deduped product list from %d to %d items", len(s.Products), len(newProducts))
	}
	s.Products = newProducts

	return nil
}

func (s *State) Looper(stream chan models.Announcement, class string, name string, interval int, fn models.LatestVersionFunction) {
	p, _ := s.GetProduct(class, name)
	log.Printf("serverpro waiting for %s version greater than %s", p, p.Latest.Version)

	for {
		maxVer, err := fn(p.Name)
		if err != nil {
			log.Printf("Error fetching %s Latest Version: %v", p, err)
		} else {
			if maxVer.GT(p.Latest.Version) {
				message := fmt.Sprintf("Version %v of %s on %s is available now", maxVer, p.Name, p.Class)
				stream <- models.Announcement{p, message}
			} else {
				message := fmt.Sprintf("Version %v of %s on %s is still the best", maxVer, p.Name, p.Class)
				// stream <- models.Announcement{p, message}
				log.Printf(message)
			}

			p.Latest.Version = maxVer
			p.Latest.Time = time.Now()

			s.PutProduct(p)
		}

		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func (s *State) LoadProducts(regFunc models.RegisterFunction, typesFunc models.GetTypesFunction) error {
	count := 0

	class, fn := regFunc()

	typeList, err := typesFunc()
	if err != nil {
		log.Printf("Error fetching %s product list: %v", class, err)
		return err
	} else {
		for _, name := range typeList {
			p := models.Product{}
			p.Name = name
			p.Class = class
			p.Function = fn

			err := s.PutProduct(p)
			if err != nil {
				log.Printf("Error putting: %v", err)
			} else {
				count++
			}
		}
	}

	log.Printf("Loaded %d products from %s (%d total)", count, class, len(s.Products))

	return nil
}
