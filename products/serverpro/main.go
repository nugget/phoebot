//
// Copyright (c) 2019 David McNett.  All Rights Reserved.
//
// SPDX-License-Identifier: BSD-2-Clause
//

package serverpro

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/nugget/phoebot/lib/config"
	"github.com/nugget/phoebot/lib/ipc"
	"github.com/nugget/phoebot/lib/products"
	"github.com/nugget/phoebot/models"

	"github.com/blang/semver"
	"github.com/sirupsen/logrus"
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
		logrus.WithFields(logrus.Fields{
			"URI":       URI,
			"bytes":     len(xmlCache),
			"cacheTime": cacheTime,
		}).Trace("Using cached server.pro gametypes")
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

		logrus.WithFields(logrus.Fields{
			"URI":   URI,
			"bytes": len(xmlCache),
		}).Info("Fetched server.pro gametypes")
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
	latestVersion := semver.MustParse("0.0.0")

	body, err := getXML()
	if err != nil {
		return latestVersion, err
	}

	server := gjson.Get(body, fmt.Sprintf("mc.%s", serverType))

	server.ForEach(func(key, value gjson.Result) bool {
		keyString := strings.Split(key.String(), " ")[0]
		logrus.WithField("keyString", keyString).Trace("evaluating version")

		v, err := semver.ParseTolerant(keyString)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"serverType": serverType,
				"err":        err,
				"key":        key.String(),
				"keyString":  keyString,
			}).Debug("ParseTolerant failed")

			return false
		}

		if v.GT(latestVersion) {
			latestVersion = v
		}

		return true
	})

	logrus.WithField("latestVersion", latestVersion).Trace("LatestVersion exiting")
	return latestVersion, nil
}

func UpdateAllProducts() error {
	body, err := getXML()
	if err != nil {
		return err
	}

	plist := gjson.Get(body, "mc")

	recordCount := 0
	updateCount := 0
	newVerCount := 0

	plist.ForEach(func(key, value gjson.Result) bool {
		recordCount++

		oldVersion, err := products.GetProduct(CLASS, key.String())
		logrus.WithFields(logrus.Fields{
			"oldVersion": oldVersion,
			"error":      err,
		}).Trace("UAV Foreach OldVersion")

		p := models.Product{}

		p.Class = CLASS
		p.Name = key.String()
		p.Latest.Time = time.Now()

		p.Latest.Version, err = LatestVersion(p.Name)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"err":   err,
				"key":   key,
				"value": value,
			}).Error("UAV LatestVersion failed")
			return false
		}

		if p.Latest.Version.GT(oldVersion.Latest.Version) {
			newVerCount++
			logrus.WithField("p", p).Trace("Sending announcement")
			ipc.AnnounceStream <- p
		}

		err = products.PutProduct(p)
		if err == nil {
			updateCount++
		} else {
			logrus.WithFields(logrus.Fields{
				"err": err,
				"key": key,
			}).Error("UAV PutProduct failed")

			return false
		}

		return true
	})

	logrus.WithFields(logrus.Fields{
		"recordCount": recordCount,
		"updateCount": updateCount,
		"newVerCount": newVerCount,
	}).Info(fmt.Sprintf("%s UpdateAllProducts", CLASS))

	return nil
}

func Poller(interval int) {
	slew := rand.Intn(10)
	interval = interval + slew

	configFlag := fmt.Sprintf("enable_%s_poller", CLASS)

	for {
		enabled, _ := config.GetBool(configFlag, false)
		if !enabled {
			logrus.Tracef("%s config is false", configFlag)
			time.Sleep(time.Duration(300) * time.Second)
			continue
		}

		logrus.WithField("interval", interval).Debug(fmt.Sprintf("Looping %s Poller", CLASS))

		err := UpdateAllProducts()
		if err != nil {
			logrus.WithError(err).Error(fmt.Sprintf("%s UpdateAllProducts failed", CLASS))
		}
		time.Sleep(time.Duration(interval) * time.Second)
	}

}
