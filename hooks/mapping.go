package hooks

import (
	"regexp"

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

	return nil
}
