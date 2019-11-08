package postal

import (
	"github.com/nugget/phoebot/lib/console"
	"github.com/nugget/phoebot/lib/db"
	"github.com/nugget/phoebot/lib/ipc"
	"github.com/nugget/phoebot/lib/phoelib"
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
}

func (m *Mailbox) Update() (bool, error) {
	newJSON, err := console.GetBlock(m.X, m.Y, m.Z, "")
	if err != nil {
		return false, err
	}

	logrus.WithFields(logrus.Fields{
		"id":   m.ID,
		"x":    m.X,
		"y":    m.Y,
		"z":    m.Z,
		"json": m.json,
		"new":  newJSON,
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
	w := models.GameWhisper{m.Name, "Mailbox contents have changed"}

	if ipc.GameWhisperStream != nil {
		ipc.GameWhisperStream <- w
	}

	return nil
}

func PollContainers() error {
	query := `SELECT containerID, x, y, z, name, json
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

	for x := sx; x <= fx; x++ {
		for y := sy; y <= fy; y++ {
			for z := sz; z <= fz; z++ {
				data, err := console.GetBlock(x, y, z, "")
				if err != nil {
					return err
				}

				logrus.WithFields(logrus.Fields{
					"x":    x,
					"y":    y,
					"z":    z,
					"data": data,
				}).Info("Container")
			}
		}
	}

	return nil
}
