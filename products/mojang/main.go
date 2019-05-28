package mojang

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/nugget/phoebot/lib/db"
	"github.com/nugget/phoebot/lib/ipc"
	"github.com/nugget/phoebot/lib/phoelib"
	"github.com/nugget/phoebot/models"

	"github.com/blang/semver"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

const CLASS = "mojang"
const URI = "https://www.minecraft.net"
const PATH = "/content/minecraft-net/_jcr_content.articles.grid"
const QUERY = "tileselection=auto&tagsPath=minecraft:article/news&propResPath=/conf/minecraft/settings/wcm/policies/minecraft/components/content/grid/policy_grid&count=2000&pageSize=20&tag=ALL&lang=/content/minecraft-net/language-masters/en-us"

func GetTypes() ([]string, error) {
	return []string{"release", "snapshot"}, nil
}

func GetHTTPBody(url string) (string, error) {
	r, err := http.Get(url)
	if err != nil {
		return "", err
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", err
	}

	body := string(bodyBytes)

	return body, nil
}

func ParseReleaseArticle(title, url string) (bool, string, string, error) {
	body, err := GetHTTPBody(url)
	if err != nil {
		return false, "", "", err
	}

	if !strings.Contains(body, "Minecraft server jar") {
		logrus.WithField("url", url).Trace("Article does not contain server.jar")
		return false, "news", "", nil
	}

	rV := regexp.MustCompile("([0-9].*)")
	if rV.MatchString(title) {
		res := rV.FindStringSubmatch(title)
		//phoelib.DebugSlice(res)

		versionString := TransformVersion(res[1])

		product := "release"
		_, err := semver.ParseTolerant(versionString)
		if err != nil {
			product = "snapshot"
		}

		return true, product, versionString, nil
	}

	return false, "unknown", "", nil
}

func TransformVersion(orig string) (new string) {
	new = strings.ToLower(orig)
	new = strings.TrimSpace(new)

	if new != orig {
		logrus.WithFields(logrus.Fields{
			"orig": orig,
			"new":  new,
		}).Trace("Transformed Mojang Version")
	}

	return new
}

func ArticleExistsInDB(a models.Article) (bool, error) {
	query := `SELECT articleID FROM mojangnews
	           WHERE title ILIKE $1 AND url = $2 AND publishdate = $3`

	rows, err := db.DB.Query(query, a.Title, a.URL, a.PublishDate)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	matchCount := 0
	for rows.Next() {
		matchCount++
	}

	if matchCount == 1 {
		// We have already logged this article
		return true, nil
	} else if matchCount > 1 {
		return true, fmt.Errorf("Unexpectedly found %d matching articles", matchCount)
	}

	return false, nil
}

func UpdateArticle(a models.Article) error {
	exists, err := ArticleExistsInDB(a)
	if err != nil {
		return err
	}

	if exists {
		logrus.WithFields(logrus.Fields{
			"title":   a.Title,
			"url":     a.URL,
			"product": a.Product,
			"release": a.Release,
			"version": a.Version,
		}).Trace("old mojang news article ignored")
		return nil
	}

	query := `INSERT INTO mojangnews (title, url, publishdate, release, product, version)
			  SELECT $1, $2, $3, $4, $5, $6
			  ON CONFLICT (title, url, publishdate) DO NOTHING`

	phoelib.LogSQL(query)
	_, err = db.DB.Exec(query, a.Title, a.URL, a.PublishDate, a.Release, a.Product, a.Version)
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"title":   a.Title,
		"url":     a.URL,
		"product": a.Product,
		"release": a.Release,
		"version": a.Version,
	}).Info("New mojang news article spotted")

	if a.Release {
		if ipc.MojangStream == nil {
			logrus.WithError(fmt.Errorf("ipc.MojangStream not initialized")).Error("Unable to send announcement")
		} else {
			logrus.WithField("a", a).Trace("Sending announcement")
			ipc.MojangStream <- a
		}
	}

	return nil
}

func ArticleFromJSON(json string) (a models.Article, err error) {
	a.URL = URI + gjson.Get(json, "article_url").String()
	a.Title = gjson.Get(json, "default_tile.title").String()

	publishedString := gjson.Get(json, "publish_date").String()
	a.PublishDate, err = time.Parse("02 Jan 2006 15:04:05 MST", publishedString)

	return a, err
}

func SeekReleases() error {
	body, err := GetHTTPBody(URI + PATH + "?" + url.QueryEscape(QUERY))
	if err != nil {
		return err
	}

	aList := gjson.Get(body, "article_grid")

	count := 0
	aList.ForEach(func(key, value gjson.Result) bool {
		count++

		aJson := value.String()

		a, err := ArticleFromJSON(aJson)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error":   err,
				"json":    aJson,
				"article": a,
			}).Error("Unable to parse article json")

		}

		a.Release, a.Product, a.Version, err = ParseReleaseArticle(a.Title, a.URL)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
				"title": a.Title,
				"url":   a.URL,
			}).Error("Unable to parse Mojang news article")
		} else {
			err := UpdateArticle(a)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err,
					"title": a.Title,
					"url":   a.URL,
				}).Error("Unable to update Mojang news article")
			}
		}

		return true
	})

	logrus.WithField("count", count).Info("Reviewed article list from mojang")

	return nil
}

func Poller(interval int) {
	slew := rand.Intn(10)
	interval = interval + slew

	for {
		logrus.WithField("interval", interval).Debug(fmt.Sprintf("Looping %s Poller", CLASS))

		err := SeekReleases()
		if err != nil {
			logrus.WithError(err).Error(fmt.Sprintf("%s SeekReleases failed", CLASS))
		}
		time.Sleep(time.Duration(interval) * time.Second)
	}

}
