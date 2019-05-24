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

const URI = "https://server.pro/r/server/getGametypes"
const CLASS = "server.pro"

var (
	xmlCache  string
	cacheTime time.Time
)

func getXML() (string, error) {
	expires := time.Now().Add(time.Duration(-60) * time.Second)

	if cacheTime.After(expires) {
		log.Printf("Using cached %d bytes from %s at %s", len(xmlCache), URI, cacheTime)
	} else {
		r, err := http.Get(URI)
		if err != nil {
			return "", err
		}

		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return "", err
		}

		xmlCache = string(bodyBytes)
		cacheTime = time.Now()

		log.Printf("Fetched %d bytes from %s at %s", len(xmlCache), URI, cacheTime)
	}

	return xmlCache, nil
}

func Register() (string, models.LatestVersionFunction) {
	return CLASS, LatestVersion
}

func GetTypes() (types []string, err error) {
	body, err := getXML()
	if err != nil {
		return types, err
	}

	plist := gjson.Get(body, "mc")

	plist.ForEach(func(key, value gjson.Result) bool {
		types = append(types, key.String())
		return true
	})

	return types, err
}

func LatestVersion(serverType string) (semver.Version, error) {
	latestVersion := semver.MustParse("0.0.1")

	body, err := getXML()
	if err != nil {
		return latestVersion, err
	}

	server := gjson.Get(body, fmt.Sprintf("mc.%s", serverType))

	server.ForEach(func(key, value gjson.Result) bool {
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
