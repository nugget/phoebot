package hooks

import (
	"fmt"
	"math/rand"
	"regexp"
	"time"

	"github.com/nugget/phoebot/lib/console"
	"github.com/nugget/phoebot/lib/ipc"
	"github.com/nugget/phoebot/lib/phoelib"
	"github.com/nugget/phoebot/lib/player"

	"github.com/sirupsen/logrus"
)

const (
	NAME     = `Phoebot's Avenger`
	PAIN     = `/effect give %s minecraft:nausea 300`
	DESPAWN  = `/kill @e[name="Phoebot's Avenger"]`
	DURATION = 300
)

func delayCommand(delay int, command string) {
	logrus.WithFields(logrus.Fields{
		"delay":   delay,
		"command": command,
	}).Debug("Delayed Command")

	if ipc.ServerSayStream == nil {
		logrus.Warn("Cannot run command, ServerSayStream is nil")
		return
	}

	time.Sleep(time.Duration(delay) * time.Second)

	ipc.ServerSayStream <- command
}

func summonRandom(player string) {
	if ipc.ServerSayStream == nil {
		logrus.Warn("Cannot summon avenger, ServerSayStream is nil")
		return
	}

	var (
		summonFmt string
		count     int
	)

	customName := fmt.Sprintf(`CustomName:"\"%s\""`, NAME)

	rand.Seed(time.Now().UnixNano())
	pick := rand.Intn(4)

	switch pick {
	case 0:
		summonFmt = `/summon vex %d %d %d {%s}`
		count = 3
	case 1:
		summonFmt = `/summon zombie %d %d %d {%s,IsBaby:1,HandItems:[{Count:1,id:golden_sword},{}],ArmorItems:[{Count:1,id:golden_boots},{Count:1,id:golden_leggings},{Count:1,id:golden_chestplate},{Count:1,id:golden_helmet}]}`
		count = 2
	case 2:
		summonFmt = `/summon piglin_brute %d %d %d {%s}`
		count = 1
	case 3:
		summonFmt = `/summon ravager %d %d %d {%s}`
		count = 1
	}

	time.Sleep(5 * time.Second)

	p, err := console.GetPlayer(player)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err":    err,
			"player": player,
		}).Error("Unable to GetPlayer")
		return
	}

	summonCmd := fmt.Sprintf(summonFmt, p.X, p.Y, p.Z, customName)

	logrus.WithFields(logrus.Fields{
		"command": summonCmd,
		"format":  summonFmt,
		"length":  len(summonCmd),
	}).Info("Executing summon")

	for i := 0; i < count; i++ {
		ipc.ServerSayStream <- summonCmd
	}

	ipc.ServerSayStream <- fmt.Sprintf(`/effect give @e[name="%s"] minecraft:regeneration %d 2`, NAME, DURATION)
	ipc.ServerSayStream <- fmt.Sprintf(`/effect give @e[name="%s"] minecraft:resistance %d 5`, NAME, DURATION)
	ipc.ServerSayStream <- fmt.Sprintf(`/effect give @e[name="%s"] minecraft:speed %d 2`, NAME, DURATION)
	ipc.ServerSayStream <- fmt.Sprintf(`/effect give @e[name="%s"] minecraft:glowing %d`, NAME, DURATION)

	return
}

func RegAvengeMe() (t Trigger) {
	t.Regexp = regexp.MustCompile("(?i)^Phoebot was (slain|killed) by (.+)")
	t.GameHook = ProcAvengeMe
	t.InGame = true

	return t
}

func ProcAvengeMe(message string) (string, error) {
	t := RegAvengeMe()
	res := t.Regexp.FindStringSubmatch(message)

	if len(res) != 3 {
		logrus.Warn(message)
		phoelib.DebugSlice(res)
		return "Can't Parse Death Message", nil
	}

	who := res[2]

	p, err := player.GetPlayerFromMinecraftName("MacNugget")
	if err != nil {
		logrus.WithError(err).Error("Cannot get MacNugget player info")
	} else {
		p.SendMessage(message)

	}

	//painCmd := fmt.Sprintf(PAIN, who)

	if ipc.ServerSayStream != nil {
		ipc.ServerSayStream <- "I have been slain!  Avenge my death!"
		ipc.ServerSayStream <- `/gamemode spectator Phoebot`
		go summonRandom(who)
	}

	return "Summoned Avenger", nil
}

func RegDespawnAvengers() (t Trigger) {
	t.Regexp = regexp.MustCompile("(?i)was (slain|killed) by Phoebot's Avenger")
	t.GameHook = ProcDespawnAvengers
	t.InGame = true

	return t
}

func ProcDespawnAvengers(message string) (string, error) {
	delayCommand(10, DESPAWN)

	return "Killed Avengers", nil
}
