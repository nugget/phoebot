package hooks

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/nugget/phoebot/lib/mcserver"
	"github.com/nugget/phoebot/lib/merchant"
	"github.com/nugget/phoebot/lib/phoelib"
	"github.com/sirupsen/logrus"
)

func RegMerchantContainer() (t Trigger) {
	t.Regexp = regexp.MustCompile(`(?i)forsale "([^"]+)" (\d+) (\d+) (\d+) (\d+) (\d+) (\d+)`)
	t.GameHook = ProcMerchantContainer
	t.InGame = true

	return t
}

func ProcMerchantContainer(message string) (string, error) {
	t := RegMerchantContainer()
	res := t.Regexp.FindStringSubmatch(message)

	who, err := mcserver.GetPlayerNameFromWhisper(message)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
			"who": who,
		}).Error("Unable to GetPlayerNameFromWhisper")
		return "", err
	}

	var coords [6]int

	for i := 0; i <= 5; i++ {
		pos := i + 2
		val, err := strconv.Atoi(res[pos])
		if err != nil {
			return fmt.Sprintf("'%s' doesn't look like a coordinate", res[pos]), err
		}

		coords[i] = val
	}

	name := res[1]

	sx := coords[0]
	sy := coords[1]
	sz := coords[2]
	fx := coords[3]
	fy := coords[4]
	fz := coords[5]

	size := phoelib.SizeOf(sx, sy, sz, fx, fy, fz)

	logrus.WithFields(logrus.Fields{
		"player": who,
		"name":   name,
		"start":  fmt.Sprintf("(%d, %d, %d)", sx, sy, sz),
		"finish": fmt.Sprintf("(%d, %d, %d)", fx, fy, fz),
		"size":   size,
	}).Info("Player requested new merchant scanrange")

	if size > 64 {
		return "That range is way to big!", fmt.Errorf("Range Too Large")
	}

	err = merchant.NewScanRange(who, name, sx, sy, sz, fx, fy, fz)
	if err != nil {
		return "I can't do that, sorry", err
	}

	return "You got it, boss!", nil
}
