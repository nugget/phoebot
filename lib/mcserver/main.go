package mcserver

import (
	"fmt"

	"github.com/Tnze/go-mc/authenticate"
	"github.com/Tnze/go-mc/bot"
	"github.com/Tnze/go-mc/chat"
	"github.com/sirupsen/logrus"
)

var (
	Client *bot.Client
)

func logFields(f logrus.Fields) logrus.Fields {
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

	return nil
}

func CleanString(c chat.Message) (buf string) {
	for _, p := range c.Extra {
		buf += p.Text
	}

	return buf
}

func Handler() error {
	err := Client.HandleGame()
	return err
}

func OnGameStart() error {
	logrus.WithFields(logFields(nil)).Info("OnGameStart")
	return nil //if err isn't nil, HandleGame() will return it.
}

func OnChatMsg(c chat.Message, pos byte) error {
	fmt.Printf("c: %+v\n", c)
	fmt.Printf("String: %+v\n", c.String())
	fmt.Printf("CleanString: %+v\n", CleanString(c))
	fmt.Printf("Text: %+v\n", c.Text)
	fmt.Printf("With: %+v\n", c.With)
	fmt.Printf("Extra: %+v\n", c.Extra)

	logrus.WithFields(logFields(logrus.Fields{
		"message": c,
		"pos":     pos,
		"string":  c.String(),
	})).Info("OnChatMsg")
	return nil
}

func OnDisconnect(c chat.Message) error {
	logrus.WithFields(logFields(logrus.Fields{
		"message": c,
	})).Info("OnDisconnect")
	return nil
}

func OnPluginMessage(channel string, data []byte) error {
	logrus.WithFields(logFields(logrus.Fields{
		"channel": channel,
		"data":    data,
	})).Info("OnPluginMessage")
	return nil
}

func OnDieMessage() error {
	logrus.WithFields(logFields(nil)).Info("DieMessage")
	return nil
}
