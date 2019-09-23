package hooks

import (
	"fmt"
	"regexp"

	"github.com/nugget/phoebot/lib/console"
	"github.com/nugget/phoebot/lib/mapping"
	"github.com/nugget/phoebot/lib/mcserver"

	"github.com/sirupsen/logrus"
)

func RegNearestPortal() (t Trigger) {
	t.Regexp = regexp.MustCompile("(?i)nearest portal")
	t.GameHook = ProcNearestPortal
	t.InGame = true

	return t
}

func ProcNearestPortal(message string) (string, error) {
	// Uncomment these lines if you need to pull out substrings from
	// the original hook regular expression.
	//
	// t := RegTemplate()
	//res := t.Regexp.FindStringSubmatch(dm.Content)
	//

	who, err := mcserver.GetPlayerNameFromWhisper(message)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
			"who": who,
		}).Error("Unable to GetPlayerNameFromWhisper")
		return "", err
	}

	p, err := console.GetPlayer(who)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
			"who": who,
		}).Error("Unable to GetPlayer")
		return "", err
	}

	poi, err := mapping.NearestPOI("portal", p.X, p.Z, p.Dimension)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
			"who": who,
		}).Error("Unable to NearestPOI")
		return "", err
	}

	fmt.Printf("BC NP: %+v", poi)

	return fmt.Sprintf("%s is %d blocks away at (%d, %d, %d)", poi, poi.Distance, poi.X, poi.Y, poi.Z), nil
}
