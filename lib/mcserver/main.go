package mcserver

import (
	"fmt"
	"time"

	"github.com/Tnze/go-mc/authenticate"
	"github.com/Tnze/go-mc/bot"
	"github.com/Tnze/go-mc/chat"
	"github.com/sirupsen/logrus"
)

var (
	Client     *bot.Client
	remoteHost string
	remotePort int
)

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
	interval := 10
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
		logrus.WithError(err).Debug("Heartbeat Loop")
		retries++

		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func CleanString(c chat.Message) (buf string) {
	for _, p := range c.Extra {
		buf += p.Text
	}

	return buf
}

func Handler() error {
	logrus.Debug("Minecraft Handler Launched")
	err := Client.HandleGame()
	logrus.WithError(err).Error("Minecraft Handler Exited")
	return err
}

func OnGameStart() error {
	logrus.WithFields(LogFields(nil)).Info("Minecraft start")
	return nil //if err isn't nil, HandleGame() will return it.
}

func OnChatMsg(c chat.Message, pos byte) error {
	// fancyMessage includes all the ANSI color codes
	fancyMessage := fmt.Sprintf("%v", c.String())
	// cleanMessage is just a 7 bit clean ASCII message
	cleanMessage := CleanString(c)

	//fmt.Printf("fancy: %+v\n", fancyMessage)
	//fmt.Printf("clean: %+v\n", cleanMessage)
	//fmt.Printf("1: %v\n2: %v\n", []byte(fancyMessage), []byte(cleanMessage))

	f := LogFields(logrus.Fields{
		"pos":   pos,
		"event": "chat",
	})

	if fancyMessage == cleanMessage {
		logrus.WithFields(f).Debug(cleanMessage)
	} else {
		logrus.WithFields(f).Info(cleanMessage)
	}

	return nil
}

func OnDisconnect(c chat.Message) error {
	logrus.WithFields(LogFields(logrus.Fields{
		"message": CleanString(c),
	})).Info("Minecraft disconnect")

	Reconnect()

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
