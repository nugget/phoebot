package mcserver

import (
	"fmt"
	"strings"
	"time"

	"github.com/Tnze/go-mc/authenticate"
	"github.com/Tnze/go-mc/bot"
	"github.com/Tnze/go-mc/chat"
	"github.com/Tnze/go-mc/data"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

var (
	Client     *bot.Client
	remoteHost string
	remotePort int
)

type PingStats struct {
	Delay         time.Duration
	PlayersOnline int64
	PlayersMax    int64
	Description   string
	Version       string
	Protocol      int64
}

func LogFields(f logrus.Fields) logrus.Fields {
	if f == nil {
		f = logrus.Fields{}
	}

	conn := Client.Conn()

	f["name"] = Client.Auth.Name
	f["server"] = conn.Socket.RemoteAddr()

	return f
}

func Login(hostname string, port int, email, password string) error {
	Client = bot.NewClient()

	auth, err := authenticate.Authenticate(email, password)
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"name":  auth.SelectedProfile.Name,
		"uuid":  auth.SelectedProfile.ID,
		"token": auth.AccessToken,
	}).Debug("Authenticated with mojang")

	Client.Name, Client.Auth.UUID, Client.AsTk = auth.SelectedProfile.Name, auth.SelectedProfile.ID, auth.AccessToken

	err = Client.JoinServer(hostname, port)
	if err != nil {
		return err
	}

	remoteHost = hostname
	remotePort = port

	return nil
}

func Reconnect() {
	interval := 30
	retries := 1

	for {
		err := Client.JoinServer(remoteHost, remotePort)
		if err == nil {
			logrus.WithFields(LogFields(logrus.Fields{
				"retries": retries,
			})).Info("Minecraft reconnected")

			go Handler()

			return
		}
		logrus.WithFields(logrus.Fields{
			"error":   err,
			"retries": retries,
		}).Info("mcserver Reconnect Loop")
		retries++

		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func CleanString(m chat.Message) string {
	var msg strings.Builder

	msg.WriteString(m.Text)

	//handle translate
	if m.Translate != "" {
		args := make([]interface{}, len(m.With))
		for i, v := range m.With {
			var arg chat.Message
			arg.UnmarshalJSON(v) //ignore error
			args[i] = arg
		}

		fmt.Fprintf(&msg, data.EnUs[m.Translate], args...)
	}

	if m.Extra != nil {
		for i := range m.Extra {
			msg.WriteString(CleanString(chat.Message(m.Extra[i])))
		}
	}

	return msg.String()
}

func ChatMsgClass(m chat.Message) string {
	if m.Translate == "commands.message.display.incoming" {
		return "whisper"
	}

	if strings.HasPrefix(m.Translate, "chat.type.text") {
		return "chat"
	}

	if strings.HasPrefix(m.Translate, "gamemode") {
		return "ignore"
	}

	if strings.HasPrefix(m.Translate, "chat.type.emote") {
		return "chat"
	}

	if strings.HasPrefix(m.Translate, "chat.type.announcement") {
		return "announcement"
	}

	if strings.HasPrefix(m.Translate, "death.") {
		return "death"
	}

	text := CleanString(m)

	if strings.HasPrefix(text, "<") {
		return "chat"
	}

	if strings.HasPrefix(text, "* ") {
		return "chat"
	}

	if strings.Contains(text, "joined the game") {
		return "join"
	}

	if strings.Contains(text, "left the game") {
		return "join"
	}

	if strings.Contains(text, "went to bed") {
		return "ignore"
	}

	return "other"
}

func Handler() error {
	logrus.Debug("Minecraft Handler Launched")
	err := Client.HandleGame()
	logrus.WithError(err).Error("Minecraft Handler Exited")
	go Reconnect()
	return err
}

func GetPingStats() (ps PingStats, err error) {
	var resp []byte

	resp, ps.Delay, err = bot.PingAndList(remoteHost, remotePort)
	if err != nil {
		return ps, err
	}

	json := string(resp)

	ps.PlayersOnline = gjson.Get(json, "players.online").Int()
	ps.PlayersMax = gjson.Get(json, "players.max").Int()
	ps.Description = gjson.Get(json, "description.text").String()
	ps.Version = gjson.Get(json, "version.name").String()
	ps.Protocol = gjson.Get(json, "verison.protocol").Int()

	ps.PlayersOnline--

	return ps, nil
}

func OnGameStart() error {
	logrus.WithFields(LogFields(nil)).Info("Minecraft start")
	return nil //if err isn't nil, HandleGame() will return it.
}

func OnChatMsg(c chat.Message, pos byte) error {
	cleanMessage := CleanString(c)

	f := LogFields(logrus.Fields{
		"pos":   pos,
		"event": "chat",
	})

	logrus.WithFields(f).Info(cleanMessage)

	return nil
}

func OnDisconnect(c chat.Message) error {
	logrus.WithFields(LogFields(logrus.Fields{
		"message": CleanString(c),
	})).Info("Minecraft disconnect")

	go Reconnect()

	return nil
}

func OnPluginMessage(channel string, data []byte) error {
	logrus.WithFields(LogFields(logrus.Fields{
		"channel": channel,
		"data":    data,
		"string":  string(data),
	})).Info("Minecraft Plugin Message")
	return nil
}

func OnDieMessage() error {
	logrus.WithFields(LogFields(nil)).Info("Minecraft death")
	return nil
}
