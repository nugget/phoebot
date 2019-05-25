package papermc

import (
	"io/ioutil"
	"net/http"

	"github.com/nugget/phoebot/models"
	"github.com/sirupsen/logrus"

	"github.com/blang/semver"
	"github.com/tidwall/gjson"
)

const URI = "https://papermc.io/api/v1/"
const CLASS = "PaperMC"

func Register() (string, models.LatestVersionFunction) {
	return CLASS, LatestVersion
}

func GetTypes() ([]string, error) {
	return []string{"paper"}, nil
}

func LatestVersion(name string) (semver.Version, error) {
	latestVersion := semver.MustParse("0.0.1")

	r, err := http.Get(URI + name)
	if err != nil {
		return latestVersion, err
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return latestVersion, err
	}

	body := string(bodyBytes)
	paper := gjson.Get(body, "versions")

	paper.ForEach(func(key, value gjson.Result) bool {
		v, err := semver.ParseTolerant(value.String())
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error":   err,
				"version": value,
			}).Warn("Unable to parse PaperMC version")
		} else {
			if v.GT(latestVersion) {
				latestVersion = v
			}
		}

		return true
	})

	return latestVersion, nil
}
