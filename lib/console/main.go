package console

import (
	"fmt"

	"github.com/seeruk/minecraft-rcon/rcon"
	"github.com/sirupsen/logrus"
)

type Connection struct {
	Hostname string
	Port     int
	password string
	client   *rcon.Client
}

var s Connection

func Initialize(hostname string, port int, password string) (err error) {
	s.Hostname = hostname
	s.Port = port
	s.password = password

	s.client, err = rcon.NewClient(s.Hostname, s.Port, s.password)
	return err
}

func (c *Connection) sendCommand(command string) (string, error) {
	if c.client == nil {
		return "", fmt.Errorf("No RCON connection")
	}

	response, err := c.client.SendCommand(command)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"command": command,
			"error":   err,
		}).Error("Failed RCON Command")
	} else if response != "" {
		logrus.WithFields(logrus.Fields{
			"command":  command,
			"response": response,
		}).Debug("Successful RCON Command")
	}

	return response, err
}

func Test() {
	s.sendCommand("version")
	s.sendCommand("data get entity MacNugget")
}
