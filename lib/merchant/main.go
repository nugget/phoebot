package merchant

import (
	"fmt"

	"github.com/nugget/phoebot/lib/db"

	"github.com/nugget/phoebot/lib/containers"
	"github.com/nugget/phoebot/lib/coreprotect"
	"github.com/nugget/phoebot/lib/player"
	"github.com/sirupsen/logrus"
)

func ScanStock() error {
	ranges, err := containers.GetRanges("inventory")
	if err != nil {
		return err
	}

	for _, sr := range ranges {
		logrus.WithFields(sr.LogFields()).Trace("Scanning for merchant activity")

		l, err := coreprotect.ScanContainers(sr.Dimension, sr.LastScan,
			sr.Sx, sr.Sy, sr.Sz, sr.Fx, sr.Fy, sr.Fz)
		if err != nil {
			return err
		}

		for i, t := range l {
			// t.User = "FakeUser"

			logrus.WithFields(logrus.Fields{
				"i":          i,
				"container":  sr.Name,
				"player":     t.User,
				"owner":      sr.Owner,
				"item":       t.Material,
				"quantity":   t.Amount,
				"action":     t.Action,
				"actionCode": t.ActionCode,
			}).Info("Merchant activity")

			if t.User != sr.Owner {
				message := fmt.Sprintf("%s %s %d %s %s %s at (%d, %d, %d)",
					t.User, t.Action, t.Amount, t.Material, t.Preposition,
					sr.Name, t.X, t.Y, t.Z)

				err = player.SendMessage(sr.Owner, message)
				if err != nil {
					return err
				}
			}
		}

	}

	return nil
}

func NewScanRange(owner, name string, sx, sy, sz, fx, fy, fz int) error {
	query := `INSERT INTO scanrange (name, sx, sy, sz, fx, fy, fz, scantype, owner)
              SELECT $1, $2, $3, $4, $5, $6, $7, $8, $9`

	_, err := db.DB.Exec(query, name, sx, sy, sz, fx, fy, fz, "inventory", owner)

	return err
}
