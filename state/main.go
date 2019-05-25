package state

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"math/rand"
	"strings"
	"time"

	"github.com/nugget/phoebot/models"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
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
			s.Subscriptions = append(s.Subscriptions, sub)
			message := fmt.Sprintf("You are now subscribed to receive updates to this channel for %s releases from %s", sub.Name, sub.Class)
			s.Dg.ChannelMessageSend(sub.ChannelID, message)

			logrus.WithFields(logrus.Fields{
				"name":      sub.Name,
				"class":     sub.Class,
				"channelID": sub.ChannelID,
				"target":    sub.Target,
			}).Info("Added new subscription")
		}

		logrus.WithField("subCount", len(s.Subscriptions)).Debug("Active subscription count")
	}

	return nil
}

func (s *State) DropSubscription(sub models.Subscription) error {
	var newSubs []models.Subscription

	for _, v := range s.Subscriptions {
		if !SubscriptionsMatch(sub, v) {
			newSubs = append(newSubs, v)
		}
	}

	if len(s.Subscriptions) != len(newSubs) {
		logrus.WithFields(logrus.Fields{
			"name":      sub.Name,
			"class":     sub.Class,
			"channelID": sub.ChannelID,
			"target":    sub.Target,
		}).Info("Deopped subscription")

		message := fmt.Sprintf("You are no longer subscribed to receive updates to this channel for %s releases from %s", sub.Name, sub.Class)
		s.Dg.ChannelMessageSend(sub.ChannelID, message)
	}

	s.Subscriptions = newSubs

	logrus.WithField("subCount", len(s.Subscriptions)).Debug("Active subscription count")

	return nil
}

func (s *State) ListSubscriptions() error {
	logrus.WithField("subCount", len(s.Subscriptions)).Info("Active subscription count")
	for i, v := range s.Subscriptions {
		channel, _ := s.Dg.State.Channel(v.ChannelID)

		logrus.WithFields(logrus.Fields{
			"channelID":   v.ChannelID,
			"channelName": channel.Name,
			"class":       v.Class,
			"name":        v.Name,
			"target":      v.Target,
		}).Infof("Subscription %d", i)
	}

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
	if n.Class == "" || n.Name == "" {
		return fmt.Errorf("Can't load malformed product: %+v", n)
	}

	newProducts := []models.Product{}

	added := false
	for _, p := range s.Products {

		if p.Class == "" || p.Name == "" {
			logrus.WithFields(logrus.Fields{
				"class": p.Class,
				"name":  p.Name,
			}).Info("Skipping malformed product")
		} else if p.Class == n.Class && p.Name == n.Name {
			if n.Function == nil {
				n.Function = p.Function
			}
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

	for i, p := range s.Products {
		key := fmt.Sprintf("%v-%v", p.Class, p.Name)
		if !exists[key] {
			if p.Class == "" || p.Name == "" {
				logrus.WithFields(logrus.Fields{
					"class": p.Class,
					"name":  p.Name,
					"i":     i,
				}).Info("Skipping malformed product (Dedupe)")
			} else {
				newProducts = append(newProducts, p)
			}
		}
		exists[key] = true
	}

	if len(s.Products) != len(newProducts) {
		logrus.WithFields(logrus.Fields{
			"startCount": len(s.Products),
			"endCount":   len(newProducts),
		}).Info("Deduped product list")
	}
	s.Products = newProducts

	return nil
}

func (s *State) Looper(stream chan models.Announcement, class string, name string, interval int, fn models.LatestVersionFunction) {
	slew := rand.Intn(10)
	interval = interval + slew

	p, err := s.GetProduct(class, name)
	if err != nil {
		logrus.WithError(err).Error("Looper unable to load product")
		return
	}

	logrus.WithFields(logrus.Fields{
		"class":         p.Class,
		"name":          p.Name,
		"latestVersion": p.Latest.Version,
		"function":      fn,
	}).Info("New Looper waiting for version")

	for {
		maxVer, err := fn(p.Name)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"class":         p.Class,
				"name":          p.Name,
				"latestVersion": p.Latest.Version,
				"error":         err,
			}).Warn("Unable to fetch latest version")
		} else {
			if maxVer.GT(p.Latest.Version) {
				logrus.WithFields(logrus.Fields{
					"class":         p.Class,
					"name":          p.Name,
					"latestVersion": p.Latest.Version,
					"newVersion":    maxVer,
				}).Info("New version detected!")

				message := fmt.Sprintf("Version %v of %s on %s is available now", maxVer, p.Name, p.Class)
				stream <- models.Announcement{p, message}
			} else {
				logrus.WithFields(logrus.Fields{
					"class":         p.Class,
					"name":          p.Name,
					"latestVersion": p.Latest.Version,
					"newVersion":    maxVer,
				}).Debug("Version unchanged")

				// Uncomment this to report versions to Discord on every fetch
				// even if the version has not changed
				//
				// message := fmt.Sprintf("Version %v of %s on %s is still the best", maxVer, p.Name, p.Class)
				// stream <- models.Announcement{p, message}
			}

			p.Latest.Version = maxVer
			p.Latest.Time = time.Now()

			err := s.PutProduct(p)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"class":         p.Class,
					"name":          p.Name,
					"latestVersion": p.Latest.Version,
					"error":         err,
				}).Warn("Unable to PutProduct")
			}
		}

		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func (s *State) LoadProducts(regFunc models.RegisterFunction, typesFunc models.GetTypesFunction) error {
	count := 0

	class, fn := regFunc()

	typeList, err := typesFunc()
	if err != nil {
		// log.Printf("Error fetching %s product list: %v", class, err)
		return err
	} else {
		for _, name := range typeList {
			p := models.Product{}
			p.Name = name
			p.Class = class
			p.Function = fn

			err := s.PutProduct(p)
			if err != nil {
				logrus.WithError(err).Warn("LoadProducts unable to PutProduct")
			} else {
				count++
			}
		}
	}

	logrus.WithFields(logrus.Fields{
		"class": class,
		"count": count,
		"total": len(s.Products),
	}).Info("Loaded productlist")

	return nil
}
