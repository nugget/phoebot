package console

import (
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
		logrus.Error("No RCON connection, attempting to connect")
		var err error
		c.client, err = rcon.NewClient(s.Hostname, s.Port, s.password)
		if err != nil {
			return "", err
		}
	}

	response, err := c.client.SendCommand(command)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"command": command,
			"error":   err,
		}).Error("Failed RCON Command")

		err = c.client.Reconnect()
		if err != nil {
			logrus.WithError(err).Error("Failed RCON Reconnect")
		} else {
			response, err = c.client.SendCommand(command)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"command": command,
					"error":   err,
				}).Error("Failed RCON Command (retry)")
			}
		}
	} else if response != "" {
		logrus.WithFields(logrus.Fields{
			"command":  command,
			"response": response,
		}).Trace("Successful RCON Command")
	}

	return response, err
}

func Test() {
	s.sendCommand("version")
	s.sendCommand("data get entity MacNugget")
}
