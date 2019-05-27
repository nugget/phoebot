package products

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/nugget/phoebot/models"

	"github.com/sirupsen/logrus"
)

func GetProduct(class, name string) (models.Product, error) {
	for _, p := range s.Products {
		if strings.ToLower(p.Class) == strings.ToLower(class) {
			if strings.ToLower(p.Name) == strings.ToLower(name) {
				return p, nil
			}
		}
	}

	return models.Product{}, fmt.Errorf("Product not found")
}

func PutProduct(n models.Product) error {
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

func DedupeProducts() error {
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

func ProductPoller(stream chan models.Announcement, class string, name string, interval int, fn models.LatestVersionFunction) {
	slew := rand.Intn(10)
	interval = interval + slew

	p, err := s.GetProduct(class, name)
	if err != nil {
		logrus.WithError(err).Error("Poller unable to load product")
		return
	}

	logrus.WithFields(logrus.Fields{
		"class":         p.Class,
		"name":          p.Name,
		"latestVersion": p.Latest.Version,
	}).Info("New Poller waiting for version")

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

func LoadProducts(regFunc models.RegisterFunction, typesFunc models.GetTypesFunction) error {
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
