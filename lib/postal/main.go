package postal

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/nugget/phoebot/lib/config"
	"github.com/nugget/phoebot/lib/console"
	"github.com/nugget/phoebot/lib/containers"
	"github.com/nugget/phoebot/lib/coreprotect"
	"github.com/nugget/phoebot/lib/ipc"
	"github.com/nugget/phoebot/lib/player"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func Reset() {
	config.WriteString("lastSignScan", "1")
}

func NewSignScan() error {
	lastScan, err := config.GetTime("lastSignScan", time.Unix(1, 0))
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"lastScan": lastScan,
	}).Trace("NewSignScan")

	ll, err := coreprotect.ScanSigns("mailbox", lastScan)
	if err != nil {
		return err
	}

	for _, l := range ll {
		// {User:MacNugget Wid:1 X:-192 Y:73 Z:-182 Line1:MacNugget Line2:Mailbox Line3: Line4:}
		b, err := coreprotect.GetBlock(l.WorldID, l.X, l.Y, l.Z)
		if err != nil {
			return err
		}

		if strings.Contains(b.Material, "wall_sign") {
			// This is a wall sign, not a floor-standing sign
			//
			signText := fmt.Sprintf("%s|%s|%s|%s", l.Line1, l.Line2, l.Line3, l.Line4)

			logrus.WithFields(logrus.Fields{
				"epoch":      l.Epoch,
				"time":       l.Timestamp,
				"player":     b.User,
				"world":      b.World,
				"item":       b.Material,
				"action":     b.Action,
				"actionCode": b.ActionCode,
				"text":       signText,
				"x":          b.X,
				"y":          b.Y,
				"z":          b.Z,
				"blockdata":  b.Blockdata,
			}).Info("New tagging wall sign placed")

			c, err := b.MountedOn()
			if err != nil {
				return err
			}

			if !IsContainer(c.Material) {
				logrus.WithFields(logrus.Fields{
					"x":        c.X,
					"y":        c.Y,
					"z":        c.Z,
					"material": c.Material,
					"world":    b.World,
					"owner":    b.User,
				}).Warn("Ignoring non-container mailbox request")
			} else {
				newBox, err := NewMailbox(Mailbox{
					Class:    "mailbox",
					Owner:    b.User,
					Signtext: signText,
					World:    b.World,
					X:        c.X,
					Y:        c.Y,
					Z:        c.Z,
					Material: c.Material,
				})
				if err != nil {
					return err
				}

				if newBox.ID != uuid.Nil && newBox.New {
					logrus.WithFields(logrus.Fields{
						"ID":       newBox.ID,
						"x":        c.X,
						"y":        c.Y,
						"z":        c.Z,
						"material": c.Material,
						"world":    b.World,
						"owner":    b.User,
					}).Info("New mailbox recorded")

					fmt.Printf("newBox: %+v\n", newBox)

					err = newBox.Rename()
					if err != nil {
						logrus.WithError(err).Error("Unable to set customName on mailbox")
					}

					err = ReSign(l, b.User)
					if err != nil {
						logrus.WithError(err).Error("Unable to set text on mailbox sign")
					}

					message := fmt.Sprintf("I'll let you know if anyone puts items in your mailbox at %d %d %d.  Feel free to update the text on the sign, I won't get confused by that.", c.X, c.Y, c.Z)
					err = player.SendMessage(b.User, message)
					if err != nil {
						logrus.WithError(err).Error("Unable to send message to player")
					}

					err = player.Advancement(b.User, "phoenixcraft:phoenixcraft/gone_postal")
					if err != nil {
						logrus.WithError(err).Warn("Unable to grant advancement")
					}
				}
			}
		}

		err = config.WriteTime("lastSignScan", l.Timestamp)
		if err != nil {
			return err
		}
	}

	return nil
}

func ReSign(l coreprotect.SignLog, player string) error {
	command := fmt.Sprintf(`/data merge block %d %d %d {Text1:"",Text2:'{"text":"%s"}',Text3:"",Text4:'{"text":"\\u2709"}'}`, l.X, l.Y, l.Z, player)
	if ipc.ServerSayStream != nil {
		ipc.ServerSayStream <- command
	}

	return nil
}

func ScanRanges() error {
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
