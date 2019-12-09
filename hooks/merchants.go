package hooks

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/nugget/phoebot/lib/mcserver"
	"github.com/nugget/phoebot/lib/merchant"
	"github.com/sirupsen/logrus"
)

func RegMerchantContainer() (t Trigger) {
	t.Regexp = regexp.MustCompile(`FORSALE "([^"]+)" (\d+) (\d+) (\d+) (\d+) (\d+) (\d+)`)
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

	fmt.Printf("Coords: %+v\n", coords)

	name := res[1]
	fmt.Printf("Name: '%+v'\n", name)

	err = merchant.NewScanRange(
		who, name,
		coords[0], coords[1], coords[2],
		coords[3], coords[4], coords[5],
	)
	if err != nil {
		return "I can't do that, sorry", err
	}

	return "You got it, boss!", nil
}
