package hooks

import (
	"fmt"
	"regexp"

	"github.com/nugget/phoebot/lib/console"
	"github.com/nugget/phoebot/lib/ipc"
	"github.com/nugget/phoebot/lib/phoelib"

	"github.com/sirupsen/logrus"
)

const AVENGER = `/summon zombie %d %d %d {IsBaby:1,HandItems:[{Count:1,id:golden_sword},{}],ArmorItems:[{Count:1,id:golden_boots},{Count:1,id:golden_leggings},{Count:1,id:golden_chestplate},{Count:1,id:golden_helmet}],CustomName:"\"Phoebot's Avenger\""}`

func RegAvengeMe() (t Trigger) {
	t.Regexp = regexp.MustCompile("(?i)Phoebot was slain by (.+)")
	t.GameHook = ProcAvengeMe
	t.InGame = true

	return t
}

func ProcAvengeMe(message string) (string, error) {
	t := RegAvengeMe()
	res := t.Regexp.FindStringSubmatch(message)

	if len(res) != 2 {
		logrus.Warn(message)
		phoelib.DebugSlice(res)
		return "Can't Parse Death Message", nil
	}

	who := res[1]

	p, err := console.GetPlayer(who)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
			"who": who,
		}).Error("Unable to GetPlayer")
		return "", err
	}

	logrus.WithFields(logrus.Fields{
		"p": p,
	}).Info("Avenge Me!")

	command := fmt.Sprintf(AVENGER, p.X, p.Y, p.Z)
	if ipc.ServerSayStream != nil {
		ipc.ServerSayStream <- "Avenge my death!"
		ipc.ServerSayStream <- command
	}

	return "Summoned Avenger", nil
}
