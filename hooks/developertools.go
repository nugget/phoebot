package hooks

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/nugget/phoebot/lib/console"
	"github.com/nugget/phoebot/lib/discord"
	"github.com/nugget/phoebot/lib/ipc"
	"github.com/nugget/phoebot/lib/phoelib"
	"github.com/nugget/phoebot/lib/postal"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

func RegLoglevel() (t Trigger) {
	t.Regexp = regexp.MustCompile("(?i)set loglevel to ([A-Z]+)")
	t.Hook = ProcLoglevel
	t.Direct = true

	return t
}

func ProcLoglevel(dm *discordgo.MessageCreate) error {
	t := RegLoglevel()
	res := t.Regexp.FindStringSubmatch(dm.Content)

	phoelib.DebugSlice(res)

	reqLevel := strings.ToLower(res[1])

	oldLevel := logrus.GetLevel()

	newLevel, err := phoelib.LogLevel(reqLevel)
	if err != nil {
		logrus.WithError(err).Info("Failed to set LogLevel")
		discord.Session.ChannelMessageSend(dm.ChannelID, fmt.Sprintf("%s", err))
		return err
	}

	if oldLevel != newLevel {
		logrus.WithFields(logrus.Fields{
			"old": oldLevel,
			"new": newLevel,
		}).Info("Console loging level changed")
	}

	message := fmt.Sprintf("Console logging level is '%s'", reqLevel)
	discord.Session.ChannelMessageSend(dm.ChannelID, message)

	return nil
}

func RegCustomName() (t Trigger) {
	t.Regexp = regexp.MustCompile("(?i)customname ([0-9-]+) ([0-9-]+) ([0-9-]+) (.*)")
	t.GameHook = ProcCustomName
	t.InGame = true

	return t
}

func ProcCustomName(message string) (string, error) {
	t := RegCustomName()
	res := t.Regexp.FindStringSubmatch(message)

	if len(res) != 5 {
		fmt.Printf("(%d) '%v'\n", len(message), message)
		fmt.Printf("(%d) '%+v'\n", len(res), res)
		fmt.Printf("(%d) '%q'\n", len(res), res)
		return "Invalid syntax", nil
	}

	x, y, z, name := res[1], res[2], res[3], res[4]

	command := fmt.Sprintf("/data modify block %s %s %s CustomName set value '{\"text\":\"%s\"}'", x, y, z, name)
	if ipc.ServerSayStream != nil {
		ipc.ServerSayStream <- command
	}

	return "Block Updated", nil
}

func RegScanMailboxes() (t Trigger) {
	t.Regexp = regexp.MustCompile("(?i)mailboxscan")
	t.GameHook = ProcScanMailboxes
	t.InGame = true

	return t
}

func ProcScanMailboxes(message string) (string, error) {
	err := postal.SearchServer(console.Hostname())
	if err != nil {
		logrus.WithError(err).Error("postal.SearchServer failure")
		return fmt.Sprintf("%s", err), err
	}
	err = postal.PollContainers()
	if err != nil {
		logrus.WithError(err).Error("postal.PollContainers failure")
		return fmt.Sprintf("%s", err), err
	}

	return "Scanned for new mailboxes", nil
}
