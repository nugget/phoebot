package hooks

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/nugget/phoebot/lib/console"
	"github.com/nugget/phoebot/lib/mapping"
	"github.com/nugget/phoebot/lib/mcserver"

	"github.com/sirupsen/logrus"
)

func RegMapMe() (t Trigger) {
	t.Regexp = regexp.MustCompile("mapme")
	t.GameHook = ProcMapMe
	t.InGame = true

	return t
}

func ProcMapMe(message string) (string, error) {
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

	mapList, err := mapping.GetByPosition(p.X, p.Z)
	if err != nil {
		logrus.WithError(err).Error("GetByPosition Failure")
	}

	var fmtMapList []string

	for _, m := range mapList {
		logrus.WithFields(logrus.Fields{
			"x":      p.X,
			"z":      p.Z,
			"mapID":  m.MapID,
			"scale":  m.Scale,
			"leftX":  m.LeftX,
			"leftZ":  m.LeftZ,
			"rightX": m.RightX,
			"rightZ": m.RightZ,
		}).Info("Position covered by map")
		fmtMapList = append(
			fmtMapList,
			fmt.Sprintf("[Map #%d 1:%d scale]", m.MapID, m.Scale),
		)
	}

	if len(fmtMapList) == 0 {
		return fmt.Sprintf(
			"There are no registered maps which cover your current location (%d, %d)",
			p.X,
			p.Z,
		), nil
	} else {
		return fmt.Sprintf(
			"There are %d maps for (%d, %d): %s",
			len(fmtMapList),
			p.X,
			p.Z,
			strings.Join(fmtMapList, ","),
		), nil
	}

	return "", fmt.Errorf("ProcMapMe Unexpected exit")
}
