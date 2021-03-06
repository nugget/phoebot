package postal

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/nugget/phoebot/lib/coreprotect"
	"github.com/nugget/phoebot/lib/db"
	"github.com/nugget/phoebot/lib/ipc"
	"github.com/nugget/phoebot/lib/phoelib"
	"github.com/nugget/phoebot/lib/player"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var Containers = map[string]bool{
	"minecraft:barrel":      true,
	"minecraft:chest":       true,
	"minecraft:shulker_box": true,
}

type Mailbox struct {
	ID       uuid.UUID
	Added    time.Time
	Changed  time.Time
	LastScan time.Time
	MaxRowID int64
	Enabled  bool
	Class    string
	Owner    string
	Signtext string
	Material string
	World    string
	WorldID  int
	X        int
	Y        int
	Z        int
	Flag     bool
	New      bool
}

func MailboxFields() string {
	return "mailboxid, added, changed, lastscan, maxrowid, enabled, class, owner, signtext, material, world, x, y, z, flag"
}

func (b *Mailbox) RowScan(rows *sql.Rows) error {
	return rows.Scan(
		&b.ID,
		&b.Added,
		&b.Changed,
		&b.LastScan,
		&b.MaxRowID,
		&b.Enabled,
		&b.Class,
		&b.Owner,
		&b.Signtext,
		&b.Material,
		&b.World,
		&b.X,
		&b.Y,
		&b.Z,
		&b.Flag,
	)
}

func (b *Mailbox) Parse() error {
	b.WorldID = coreprotect.WidFromWorld(b.World)

	b.CaseName()

	return nil
}

func (b *Mailbox) CaseName() string {
	outbuf := b.Signtext
	outbuf = strings.ReplaceAll(outbuf, "[mailbox]", "")
	outbuf = strings.ReplaceAll(outbuf, "[shop]", "")
	outbuf = strings.ReplaceAll(outbuf, "mailbox", "")
	outbuf = strings.ReplaceAll(outbuf, "shop", "")
	outbuf = strings.ReplaceAll(outbuf, "|", " ")
	outbuf = strings.ReplaceAll(outbuf, "  ", " ")
	outbuf = strings.TrimSpace(outbuf)

	return outbuf
}

func (b *Mailbox) IsContainer() bool {
	return IsContainer(b.Material)
}

func (b *Mailbox) SetFlag(flag bool) error {
	query := `UPDATE mailbox SET flag = $1 WHERE mailboxid = $2 AND flag <> $1`
	phoelib.LogSQL(query, flag, b.ID)
	_, err := db.DB.Exec(query, flag, b.ID)
	if err != nil {
		return err
	}
	b.Flag = flag

	return nil
}

func (b *Mailbox) Scanned(maxRowID int64) error {
	query := `UPDATE mailbox SET maxrowid = $2, lastscan = current_timestamp at time zone 'utc'  WHERE mailboxid = $1`
	phoelib.LogSQL(query, b.ID, maxRowID)
	_, err := db.DB.Exec(query, b.ID, maxRowID)
	if err != nil {
		return err
	}

	return nil
}

func (b *Mailbox) Rename() error {
	customName := b.Owner + `\'s Mailbox`
	command := fmt.Sprintf(`/data modify block %d %d %d CustomName set value '{"text":"%s"}'`, b.X, b.Y, b.Z, customName)

	if ipc.ServerSayStream != nil {
		ipc.ServerSayStream <- command

		logrus.WithFields(logrus.Fields{
			"world":    b.World,
			"x":        b.X,
			"y":        b.Y,
			"z":        b.Z,
			"owner":    b.Owner,
			"material": b.Material,
			"name":     customName,
		}).Info("Renamed mailbox container")
	}

	return nil
}

func (b *Mailbox) Delete() error {
	query := `UPDATE mailbox SET deleted = current_timestamp at time zone 'utc' WHERE mailboxID = $1`
	phoelib.LogSQL(query, b.ID)
	_, err := db.DB.Exec(query, b.ID)
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"world":    b.World,
		"x":        b.X,
		"y":        b.Y,
		"z":        b.Z,
		"owner":    b.Owner,
		"material": b.Material,
	}).Info("deleted mailbox record")

	return nil
}

func IsContainer(material string) bool {
	if strings.Contains(material, "shulker_box") {
		return true
	}
	return Containers[material]
}

func NewMailbox(b Mailbox) (Mailbox, error) {
	existingBox, err := GetMailboxByOwnerAndLocation(b.Owner, b.World, b.X, b.Y, b.Z)
	if err != nil {
		return Mailbox{}, err
	}

	logAction := "unknown Mailbox operation"

	if existingBox.ID != uuid.Nil {
		if b.Owner == existingBox.Owner {
			existingBox.Signtext = b.Signtext
			existingBox.New = true

			query := `UPDATE mailbox SET signtext = $2 WHERE mailboxid = $1`
			_, err := db.DB.Exec(query, existingBox.ID, existingBox.Signtext)
			if err != nil {
				return Mailbox{}, err
			}
			logAction = "Renamed existing mailbox"
		} else {
			logAction = "Ignored existing mailbox"
		}

		logrus.WithFields(logrus.Fields{
			"id":       existingBox.ID,
			"material": existingBox.Material,
			"x":        existingBox.X,
			"y":        existingBox.Y,
			"z":        existingBox.Z,
			"owner":    existingBox.Owner,
			"signtext": existingBox.Signtext,
		}).Info(logAction)

		return existingBox, nil
	}

	query := `INSERT INTO mailbox (class, owner, signtext, material, world, x, y, z)
			  SELECT $1, $2, $3, $4, $5, $6, $7, $8 RETURNING ` + MailboxFields()

	phoelib.LogSQL(query, b.Class, b.Owner, b.Signtext, b.Material, b.World, b.X, b.Y, b.Z)
	rows, err := db.DB.Query(query, b.Class, b.Owner, b.Signtext, b.Material, b.World, b.X, b.Y, b.Z)
	if err != nil {
		return Mailbox{}, err
	}
	defer rows.Close()

	for rows.Next() {
		err = b.RowScan(rows)
		if err != nil {
			return Mailbox{}, err
		}
		b.Parse()
		b.New = true
		return b, nil
	}

	return Mailbox{}, nil
}

func GetMailboxByOwnerAndLocation(owner, world string, x, y, z int) (Mailbox, error) {
	query := `SELECT ` + MailboxFields() + ` FROM mailbox WHERE deleted IS NULL AND owner = $1 AND world = $2
	AND x = $3 AND y = $4 AND z = $5`

	phoelib.LogSQL(query, owner, world, x, y, z)
	rows, err := db.DB.Query(query, owner, world, x, y, z)

	if err != nil {
		return Mailbox{}, err
	}
	defer rows.Close()

	for rows.Next() {
		b := Mailbox{}
		err = b.RowScan(rows)
		if err != nil {
			return Mailbox{}, err
		}
		b.Parse()
		return b, nil
	}
	return Mailbox{}, nil
}

func PollMailboxes() error {
	ll, err := GetMailboxes()
	if err != nil {
		return err
	}

	for _, l := range ll {
		err := PollMailbox(l)
		if err != nil {
			logrus.WithError(err).Error("PollMailbox Failure")
			continue
		}
	}

	return nil
}

func PollMailbox(l Mailbox) error {
	var message string

	l.Parse()

	// Check for destruction
	b, err := coreprotect.GetBlock(l.WorldID, l.X, l.Y, l.Z)
	if err != nil {
		return err
	}

	destroyed := false

	if b.Action == "destroyed" {
		// block log action is literally destroyed
		destroyed = true
	}

	if b.Material != "" && b.Material != l.Material {
		// New block material is not the same as old block material
		destroyed = true
	}

	if destroyed {
		logrus.WithFields(logrus.Fields{
			"action":      b.Action,
			"material":    b.Material,
			"oldMaterial": l.Material,
		}).Trace("Mailbox destruction detection")

		err := l.Delete()
		if err != nil {
			logrus.WithError(err).Error("Mailbox was destroyed, unable to remove record")
			return err
		}

		logrus.Warn("Mailbox was destroyed, removed record")

		switch l.Class {
		case "mailbox":
			message = fmt.Sprintf("Your mailbox at %d %d %d seems to be gone, so I will stop checking it.", l.X, l.Y, l.Z)
		case "shop":
			message = fmt.Sprintf("Your display case at %d %d %d seems to be gone, so I will stop checking it.", l.X, l.Y, l.Z)
		}
		err = player.SendMessage(l.Owner, message)
		if err != nil {
			logrus.WithError(err).Error("Unable to send message to player")
		}

		return nil
	}

	//fmt.Printf("Poll Mailbox: %+v\n", l)
	//fmt.Printf("Poll Container:  %+v\n", b)

	ll, err := coreprotect.ContainerActivity(l.World, l.MaxRowID, l.X, l.Y, l.Z)
	if err != nil {
		return err
	}

	for i, t := range ll {
		logrus.WithFields(logrus.Fields{
			"i":          i,
			"player":     t.Player,
			"owner":      l.Owner,
			"item":       t.Material,
			"quantity":   t.Amount,
			"action":     t.Action,
			"actionCode": t.ActionCode,
			"x":          t.X,
			"y":          t.Y,
			"z":          t.Z,
			"maxRow":     t.MaxRowID,
		}).Info("Mailbox activity")

		if l.X != t.X || l.Y != t.Y || l.Z != t.Z {
			logrus.WithFields(logrus.Fields{
				"i":          i,
				"player":     t.Player,
				"owner":      l.Owner,
				"item":       t.Material,
				"quantity":   t.Amount,
				"action":     t.Action,
				"actionCode": t.ActionCode,
				"cpX":        t.X,
				"cpY":        t.Y,
				"cpZ":        t.Z,
				"pbX":        l.X,
				"pbY":        l.Y,
				"pbZ":        l.Z,
				"maxRow":     t.MaxRowID,
			}).Error("Bad result from ContainerActivity")
		}

		if l.Owner != "" {
			if t.Player == l.Owner {
				if l.Class == "mailbox" && t.Action == "took" && l.Flag {
					err = player.Advancement(b.User, "phoenixcraft:phoenixcraft/ygm")
					if err != nil {
						logrus.WithError(err).Warn("Unable to grant advancement")
					}
				}
				l.SetFlag(false)
			} else {
				switch l.Class {
				case "mailbox":
					message = fmt.Sprintf("Someone %s items %s your mailbox at (%d, %d, %d)",
						t.Action, t.Preposition, t.X, t.Y, t.Z)

					if t.Action == "placed" {
						err = player.Advancement(b.User, "phoenixcraft:phoenixcraft/return_address")
						if err != nil {
							logrus.WithError(err).Warn("Unable to grant advancement")
						}
					}
				case "shop":
					message = fmt.Sprintf("%s %s %d %s %s your display case at (%d, %d, %d) [%s]",
						t.Player, t.Action, t.Amount, t.Material, t.Preposition, t.X, t.Y, t.Z, l.CaseName())

					if t.Action == "placed" && t.Material == "Diamond" {
						err = player.Advancement(b.User, "phoenixcraft:phoenixcraft/spendthrift")
						if err != nil {
							logrus.WithError(err).Warn("Unable to grant advancement")
						}
					}
				}

				l.SetFlag(true)

				err = player.SendMessage(l.Owner, message)
				if err != nil {
					logrus.WithError(err).Warn("Unable to SendMessage player")
				}
			}
		}

		err = l.Scanned(t.MaxRowID)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetMailboxes() (ll []Mailbox, err error) {
	query := `SELECT ` + MailboxFields() + ` from mailbox WHERE deleted IS NULL AND enabled IS TRUE`

	phoelib.LogSQL(query)
	rows, err := db.DB.Query(query)
	if err != nil {
		return ll, err
	}
	defer rows.Close()

	for rows.Next() {
		b := Mailbox{}
		err = b.RowScan(rows)
		if err != nil {
			return ll, err
		}
		b.Parse()

		ll = append(ll, b)
	}

	return ll, nil
}
