package hooks

import (
	"regexp"

	"github.com/nugget/phoebot/lib/console"

	"github.com/sirupsen/logrus"
)

func RegMapMe() (t Trigger) {
	t.Regexp = regexp.MustCompile("mapme")
	t.GameHook = ProcMapMe
	t.InGame = true

	return t
}

func ProcMapMe(message string) error {
	// Uncomment these lines if you need to pull out substrings from
	// the original hook regular expression.
	//
	// t := RegTemplate()
	//res := t.Regexp.FindStringSubmatch(dm.Content)
	//

	logrus.Warn("Hey!")

	p, err := console.GetPlayer("MacNugget")
	logrus.WithFields(logrus.Fields{
		"p":   p,
		"err": err,
	}).Info("GetPlayer from ProcMapMe")

	return nil
}
