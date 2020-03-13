package products

import (
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	"github.com/nugget/phoebot/lib/config"
	"github.com/nugget/phoebot/lib/db"
	"github.com/nugget/phoebot/lib/ipc"
	"github.com/nugget/phoebot/lib/phoelib"
	"github.com/nugget/phoebot/models"

	"github.com/blang/semver"
	"github.com/sirupsen/logrus"
)

type ProductRec struct {
	Class   string
	Name    string
	Version string
	Changed time.Time
}

func (r *ProductRec) Scan(rows *sql.Rows) error {
	err := rows.Scan(
		&r.Class,
		&r.Name,
		&r.Version,
		&r.Changed,
	)

	return err
}

func ProductRecToStruct(buf ProductRec) (p models.Product, err error) {
	p.Name = buf.Name
	p.Class = buf.Class
	p.Latest.Version, err = semver.Parse(buf.Version)
	if err != nil {
		return p, err
	}
	p.Latest.Time = buf.Changed

	return p, nil
}

func GetProduct(class, name string) (p models.Product, err error) {
	query := `SELECT class, name, version, changed FROM product
			  WHERE deleted IS NULL AND
			        class ILIKE $1 AND name ILIKE $2`

	phoelib.LogSQL(query, class, name)
	rows, err := db.DB.Query(query, class, name)
	if err != nil {
		return p, err
	}
	defer rows.Close()

	for rows.Next() {
		r := ProductRec{}

		err := r.Scan(rows)
		if err != nil {
			fmt.Printf("Yeah this is where I broke.\n")
			return p, err
		}

		p, err = ProductRecToStruct(r)
		return p, err
	}

	return p, fmt.Errorf("Product not found")
}

func GetImportant() (pList []models.Product, err error) {
	cutoff := semver.MustParse("0.0.0")

	query := `SELECT class, name, version, changed FROM product
			  WHERE deleted IS NULL`

	rows, err := db.DB.Query(query)

	if err != nil {
		return pList, err
	}
	defer rows.Close()

	for rows.Next() {
		r := ProductRec{}

		err := r.Scan(rows)
		if err != nil {
			return pList, err
		}

		p, err := ProductRecToStruct(r)
		if err != nil {
			return pList, err
		}

		if p.Latest.Version.GT(cutoff) {
			pList = append(pList, p)
		}
	}

	return pList, nil

}

func PutProduct(n models.Product) error {
	var query string

	if n.Class == "" || n.Name == "" {
		return fmt.Errorf("Can't load malformed product: %+v", n)
	}

	versionString := fmt.Sprintf("%s", n.Latest.Version)

	p, _ := GetProduct(n.Class, n.Name)
	if p.Name != "" {
		query = `UPDATE product SET version = $3 WHERE class ILIKE $1 AND name ILIKE $2`
	} else {
		query = `INSERT INTO product (class, name, version) SELECT $1,$2,$3`
	}

	phoelib.LogSQL(query, n.Class, n.Name, versionString)
	_, err := db.DB.Exec(query, n.Class, n.Name, versionString)
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"class":   n.Class,
		"name":    n.Name,
		"version": versionString,
	}).Trace("PutProduct Successful")

	return nil
}

func Poller(class string, name string, interval int, fn models.LatestVersionFunction) {
	slew := rand.Intn(10)
	interval = interval + slew

	p, err := GetProduct(class, name)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"class": class,
			"name":  name,
		}).Error("Poller unable to load product")
		return
	}

	logrus.WithFields(logrus.Fields{
		"class":         p.Class,
		"name":          p.Name,
		"latestVersion": p.Latest.Version,
	}).Info("New Poller waiting for version")

	for {
		enabled, _ := config.GetBool("enable_product_poller", false)
		if !enabled {
			logrus.Trace("enable_product_poller config is false")
			time.Sleep(time.Duration(300) * time.Second)
			continue
		}

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

				p.Latest.Version = maxVer
				p.Latest.Time = time.Now()

				ipc.AnnounceStream <- p
			} else {
				logrus.WithFields(logrus.Fields{
					"class":         p.Class,
					"name":          p.Name,
					"latestVersion": p.Latest.Version,
					"newVersion":    maxVer,
				}).Trace("Version unchanged")

				// Uncomment this to report versions to Discord on every fetch
				// even if the version has not changed
				//
				// message := fmt.Sprintf("Version %v of %s on %s is still the best", maxVer, p.Name, p.Class)
				// ipc.AnnounceStream <- models.Announcement{p, message}
			}

			err := PutProduct(p)
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

			err := PutProduct(p)
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
	}).Info("Loaded productlist")

	return nil
}
