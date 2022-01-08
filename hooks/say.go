package hooks

import (
	"regexp"
	"strings"

	"github.com/nugget/phoebot/lib/discord"
	"github.com/nugget/phoebot/lib/ipc"
	"github.com/nugget/phoebot/lib/phoelib"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

func RegSay() (t Trigger) {
	t.Regexp = regexp.MustCompile("(?i)(say) ([^ ]+) ([^ ]+) (.*)")
	t.Hook = ProcSay
	t.Direct = true
	t.ACL = "admin"

	return t
}

func ProcSay(dm *discordgo.MessageCreate) error {
	var err error

	t := RegSay()
	res := t.Regexp.FindStringSubmatch(dm.Content)

	phoelib.DebugSlice(res)

	if len(res) == 5 {
		targetType := strings.ToLower(res[2])
		target := strings.ToLower(res[3])
		message := res[4]

		switch targetType {
		case "channel":
			err = sendToChannel(target, message)
		case "server":
			err = sendToServer(target, message)
		default:
			logrus.WithFields(logrus.Fields{
				"targetType": targetType,
				"target":     target,
				"message":    message,
			}).Warn("Unknown SAY targetType")
			err = nil
		}
	}

	return err
}

func sendToChannel(target, message string) error {
	c, err := discord.GetChannelByName(target)
	if err != nil {
		return err
	}

	_, err = discord.Session.ChannelMessageSend(c.ID, message)
	return err
}

// say server _ message goes here
func sendToServer(target, message string) error {
	if ipc.ServerSayStream != nil {
		ipc.ServerSayStream <- message
	}

	return nil
}
