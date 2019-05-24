//
// Copyright (c) 2019 David McNett.  All Rights Reserved.
//
// SPDX-License-Identifier: BSD-2-Clause
//

package serverpro

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/nugget/phoebot/models"

	"github.com/blang/semver"
	"github.com/tidwall/gjson"
)

const url = "https://server.pro/r/server/getGametypes"

var (
	xmlCache  string
	cacheTime time.Time
)

func getXML() (string, error) {
	r, err := http.Get(url)
	if err != nil {
		return "", err
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", err
	}

	xmlCache := string(bodyBytes)
	cacheTime := time.Now()

	log.Printf("Fetched %d bytes from %d at %s", len(xmlCache), url, cacheTime)

	return xmlCache, nil
}

func Register() (string, models.LatestVersionFunction) {
	return "hosted", LatestVersion
}

func GetTypes() (types []string, err error) {
	return []string{"Paper"}, nil
}

func LatestVersion(serverType string) (semver.Version, error) {
	latestVersion := semver.MustParse("0.0.1")

	body, err := getXML()
	if err != nil {
		return latestVersion, err
	}

	paper := gjson.Get(body, fmt.Sprintf("mc.%s", serverType))

	paper.ForEach(func(key, value gjson.Result) bool {
		v, err := semver.ParseTolerant(key.String())
		if err != nil {
			return false
		}

		if v.GT(latestVersion) {
			latestVersion = v
		}

		return true
	})

	return latestVersion, nil
}
