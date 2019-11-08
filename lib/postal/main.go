package postal

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/nugget/phoebot/lib/console"
	"github.com/nugget/phoebot/lib/db"
	"github.com/nugget/phoebot/lib/discord"
	"github.com/nugget/phoebot/lib/ipc"
	"github.com/nugget/phoebot/lib/phoelib"
	"github.com/nugget/phoebot/lib/player"
	"github.com/nugget/phoebot/models"

	"github.com/sirupsen/logrus"
)

type Slot struct {
	Count int
	ID    string
}

type Mailbox struct {
	ID        string
	X         int
	Y         int
	Z         int
	Dimension string
	json      string
	Name      string
	Items     []string
	PlayerID  string
}

func (m *Mailbox) Update() (bool, error) {
	newJSON, err := console.GetBlock(m.X, m.Y, m.Z, "")
	if err != nil {
		return false, err
	}

	logrus.WithFields(logrus.Fields{
		"id":      m.ID,
		"x":       m.X,
		"y":       m.Y,
		"z":       m.Z,
		"json":    m.json,
		"new":     newJSON,
		"changed": (m.json != newJSON),
	}).Info("UpdateContainer")

	if m.json == newJSON {
		return false, nil
	}

	query := `UPDATE container SET json = $2 WHERE containerID = $1`

	phoelib.LogSQL(query, m.ID, newJSON)
	_, err = db.DB.Exec(query, m.ID, newJSON)

	return true, err
}

func (m *Mailbox) Notify() error {
	if m.PlayerID == "" {
		// No player known, skipping notification
		return errors.New("Can't notify on a mailbox with no playerID")
	}

	message := fmt.Sprintf("You've got mail in your mailbox at (%d, %d, %d)", m.X, m.Y, m.Z)

	gameNick, err := player.GameNickFromPlayerID(m.PlayerID)
	if err != nil {
		return err
	}

	if gameNick == "" {
		re := regexp.MustCompile("^([^']+)'s Mailbox")
		res := re.FindStringSubmatch(m.Name)
		gameNick = res[1]
		logrus.WithFields(logrus.Fields{
			"name":     m.Name,
			"gameNick": gameNick,
		}).Debug("No game nick from playerID, parsing name")
	}

	if gameNick != "" {
		w := models.Whisper{gameNick, message}

		if ipc.ServerWhisperStream != nil {
			ipc.ServerWhisperStream <- w
		}
	}

	channel, err := discord.GetChannelByPlayerID(m.PlayerID)
	if err != nil {
		return err
	}

	discord.Session.ChannelMessageSend(channel.ID, message)

	return nil
}

func PollContainers() error {
	query := `SELECT containerID, x, y, z, name, json, playerID
			  FROM container WHERE deleted IS NULL AND enabled IS TRUE`

	phoelib.LogSQL(query)
	rows, err := db.DB.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		m := Mailbox{}
		err := rows.Scan(
			&m.ID,
			&m.X,
			&m.Y,
			&m.Z,
			&m.Name,
			&m.json,
			&m.PlayerID,
		)
		if err != nil {
			return err
		}

		changed, err := m.Update()
		if err != nil {
			return err
		}
		if changed {
			return m.Notify()
		}
	}

	return nil
}

func IsContainer(json string) bool {
	return true
}

func PlayerNameFromJSON(json string) (string, error) {
	return "", nil
}

func SearchForMailboxes(sx, sy, sz, fx, fy, fz int) error {
	if sx > fx {
		sx, fx = fx, sx
	}

	if sy > fy {
		sy, fy = fy, sy
	}

	if sz > fz {
		sz, fz = fz, sz
	}

	logrus.WithFields(logrus.Fields{
		"sx": sx,
		"sy": sy,
		"sz": sz,
		"fx": fx,
		"fy": fy,
		"fz": fz,
	}).Info("SearchForMailboxes")

	for x := sx; x <= fx; x++ {
		for y := sy; y <= fy; y++ {
			for z := sz; z <= fz; z++ {
				data, err := console.GetBlock(x, y, z, "")
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"x":   x,
						"y":   y,
						"z":   z,
						"err": err,
					}).Info("GetBlock Failure")
					continue
				}

				if IsContainer(data) {
					playerName, err := PlayerNameFromJSON(data)
					if err != nil {
						logrus.WithError(err).Error("PlayerNameFromJSON Failed")
						continue
					}

					logrus.WithFields(logrus.Fields{
						"x":          x,
						"y":          y,
						"z":          z,
						"data":       data,
						"playerName": playerName,
					}).Info("Container")
				}
			}
		}
	}

	return nil
}
