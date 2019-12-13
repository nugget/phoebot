package hooks

import (
	"fmt"
	"regexp"
	"strconv"
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
			fmt.Sprintf("[#%d 1:%d]\n", m.MapID, m.Scale),
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
			"I found %d maps: %s",
			len(fmtMapList),
			strings.Join(fmtMapList, ","),
		), nil
	}
}

func RegNewMap() (t Trigger) {
	t.Regexp = regexp.MustCompile("(?i)newmap ([0-9]+) ([0-9])")
	t.GameHook = ProcNewMap
	t.InGame = true

	return t
}

func ProcNewMap(message string) (string, error) {
	t := RegNewMap()
	res := t.Regexp.FindStringSubmatch(message)

	mapid, err := strconv.Atoi(res[1])
	if err != nil {
		return fmt.Sprintf("'%s' doesn't look like a map ID", res[1]), err
	}
	scale, err := strconv.Atoi(res[2])
	if err != nil {
		return fmt.Sprintf("'%s' doesn't look like a map scaling", res[2]), err
	}

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

	m := mapping.NewMap()

	m.Scale = scale
	m.MapID = mapid
	m.LeftX, m.LeftZ, m.RightX, m.RightZ = mapping.MapBoundaries(p.X, p.Z, scale)
	m.Owner = who

	logrus.WithFields(logrus.Fields{
		"scale":   scale,
		"mapid":   mapid,
		"playerx": p.X,
		"playerz": p.Z,
		"leftx":   m.LeftX,
		"leftz":   m.LeftZ,
		"rightx":  m.RightX,
		"rightz":  m.RightZ,
	}).Debug("MapBoundaries")

	desc := fmt.Sprintf("Map #%d scaled 1:%d spans from (%d, %d) to (%d, %d)",
		m.MapID,
		m.Scale,
		m.LeftX, m.LeftZ,
		m.RightX, m.RightZ,
	)

	logrus.Info(desc)

	err = mapping.Update(m)
	if err != nil {
		return "I wasn't able to register your map, sorry.", err
	}

	return desc, nil
}

func RegPOI() (t Trigger) {
	t.Regexp = regexp.MustCompile("(?i)poi ([^ ]+) (.+)")
	t.GameHook = ProcPOI
	t.InGame = true

	return t
}

func ProcPOI(message string) (string, error) {
	t := RegPOI()
	res := t.Regexp.FindStringSubmatch(message)

	if len(res) != 3 || res[1] == "" || res[2] == "" {
		return fmt.Sprintf("You need to tell me what the POI is"), fmt.Errorf("Invalid POI")
	}

	poi := mapping.NewPOI()

	poi.Class = res[1]
	poi.Description = res[2]

	who, err := mcserver.GetPlayerNameFromWhisper(message)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
			"who": poi.Owner,
		}).Error("Unable to GetPlayerNameFromWhisper")
		return "", err
	}

	poi.Owner = who

	p, err := console.GetPlayer(who)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
			"who": who,
		}).Error("Unable to GetPlayer")
		return "", err
	}

	poi.X = p.X
	poi.Y = p.Y
	poi.Z = p.Z
	poi.Dimension = p.Dimension

	logrus.WithFields(logrus.Fields{
		"player":      poi.Owner,
		"class":       poi.Class,
		"description": poi.Description,
		"playerx":     poi.X,
		"playery":     poi.Y,
		"playerz":     poi.Z,
		"dimension":   poi.Dimension,
	}).Info("New POI")

	desc := fmt.Sprintf(
		"I'll remember that %s is at (%d, %d, %d)",
		poi.Description,
		poi.X,
		poi.Y,
		poi.Z,
	)

	err = poi.Update()
	if err != nil {
		return "I wasn't able to register your POI, sorry.", err
	}

	return desc, nil
}
