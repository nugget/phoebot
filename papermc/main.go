package papermc

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/nugget/phoebot/models"

	"github.com/blang/semver"
	"github.com/tidwall/gjson"
)

const url = "https://papermc.io/api/v1/"

func Register() (string, models.LatestVersionFunction) {
	return "PaperMC", LatestVersion
}

func GetTypes() ([]string, error) {
	return []string{"paper"}, nil
}

func LatestVersion(name string) (semver.Version, error) {
	latestVersion := semver.MustParse("0.0.1")

	r, err := http.Get(url + name)
	if err != nil {
		return latestVersion, err
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return latestVersion, err
	}

	body := string(bodyBytes)
	paper := gjson.Get(body, "versions")
	log.Printf("%v", paper)

	paper.ForEach(func(key, value gjson.Result) bool {
		v, err := semver.ParseTolerant(value.String())
		if err != nil {
			//log.Printf("Unable to parse version '%v': %v", value, err)
		} else {
			if v.GT(latestVersion) {
				latestVersion = v
			}
		}

		return true
	})

	return latestVersion, nil
}
