//
// Copyright (c) 2019 David McNett.  All Rights Reserved.
//
// SPDX-License-Identifier: BSD-2-Clause
//

package serverpro

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/blang/semver"
	"github.com/tidwall/gjson"
)

const url = "https://server.pro/r/server/getGametypes"

func LatestVersion(serverType string) (semver.Version, error) {
	latestVersion := semver.MustParse("0.0.1")

	r, err := http.Get(url)
	if err != nil {
		return latestVersion, err
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return latestVersion, err
	}

	body := string(bodyBytes)

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
