package merchant

import (
	"fmt"

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

	for i, sr := range ranges {
		logrus.WithFields(logrus.Fields{
			"i":           i,
			"scanRangeID": sr.ScanRangeID,
			"lastScan":    sr.LastScan,
			"dimension":   sr.Dimension,
			"name":        sr.Name,
			"owner":       sr.Owner,
			"start":       fmt.Sprintf("(%d, %d, %d)", sr.Sx, sr.Sy, sr.Sz),
			"finish":      fmt.Sprintf("(%d, %d, %d)", sr.Fx, sr.Fy, sr.Fz),
		}).Info("Scanning for merchant activity")

		l, err := coreprotect.ScanContainers(sr.Dimension, sr.LastScan,
			sr.Sx, sr.Sy, sr.Sz, sr.Fx, sr.Fy, sr.Fz)
		if err != nil {
			return err
		}

		for i, t := range l {
			if t.User == sr.Owner {
				logrus.WithFields(logrus.Fields{
					"i":         i,
					"container": sr.Name,
					"owner":     sr.Owner,
					"item":      t.Material,
					"quantity":  t.Amount,
					"action":    t.Action,
				}).Info("Owner merchant activity")
			} else {
				message := fmt.Sprintf("%s took %d %s from %s at (%d, %d, %d)",
					t.User, t.Amount, t.Material, sr.Name, t.X, t.Y, t.Z)

				err = player.SendMessage(sr.Owner, message)
				if err != nil {
					return err
				}
			}
		}

	}

	return nil
}
