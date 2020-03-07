package postal

import (
	"fmt"
	"regexp"
	"time"

	"github.com/nugget/phoebot/lib/config"
	"github.com/nugget/phoebot/lib/console"
	"github.com/nugget/phoebot/lib/containers"
	"github.com/nugget/phoebot/lib/coreprotect"
	"github.com/nugget/phoebot/lib/player"

	"github.com/sirupsen/logrus"
)

func NewSignScan() error {
	logrus.Info("NewSignScan")

	lastScan, err := config.GetTime("lastSignScan", time.Unix(1, 0))

	ll, err := coreprotect.ScanSigns("mailbox", lastScan)
	if err != nil {
		return err
	}

	for _, l := range ll {
		fmt.Printf("%+v\n", l)
		config.WriteTime("lastSignScan", l.Timestamp)
	}

	return nil

}

func ScanMailboxes() error {
	ranges, err := containers.GetRanges("mailboxes")
	if err != nil {
		return err
	}

	for _, sr := range ranges {
		logrus.WithFields(sr.LogFields()).Trace("Scanning for mailbox activity")

		l, err := coreprotect.ScanContainers(sr.Dimension, sr.LastScan,
			sr.Sx, sr.Sy, sr.Sz, sr.Fx, sr.Fy, sr.Fz)
		if err != nil {
			return err
		}

		for i, t := range l {
			if sr.Owner == "" {
				customName, err := console.GetText(t.X, t.Y, t.Z, "CustomName")
				if err != nil {
					logrus.WithError(err).Error("Failed to get Mailbox CustomName")
				} else {
					re := regexp.MustCompile("^([^']+)'s Mailbox")
					res := re.FindStringSubmatch(customName)
					if len(res) == 2 {
						logrus.Trace("Found owner from Mailbox CustomName")
						sr.Owner = res[1]
					} else {
						logrus.WithFields(logrus.Fields{
							"customName": customName,
							"x":          t.X,
							"y":          t.Y,
							"z":          t.Z,
						}).Warn("No owner in Mailbox CustomName")
					}
				}
			}

			logrus.WithFields(logrus.Fields{
				"i":          i,
				"range":      sr.Name,
				"player":     t.User,
				"owner":      sr.Owner,
				"item":       t.Material,
				"quantity":   t.Amount,
				"action":     t.Action,
				"actionCode": t.ActionCode,
				"x":          t.X,
				"y":          t.Y,
				"z":          t.Z,
			}).Info("Mailbox activity")

			if sr.Owner != "" {
				if t.User != sr.Owner {
					message := fmt.Sprintf("Someone %s items %s your %s mailbox at (%d, %d, %d)",
						t.Action, t.Preposition, sr.Name, t.X, t.Y, t.Z)

					err = player.SendMessage(sr.Owner, message)
					if err != nil {
						return err
					}
				}
			}
		}

	}

	return nil
}
