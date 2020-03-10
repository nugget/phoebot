package postal

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/nugget/phoebot/lib/coreprotect"
	"github.com/nugget/phoebot/lib/db"
	"github.com/nugget/phoebot/lib/ipc"
	"github.com/nugget/phoebot/lib/phoelib"
	"github.com/nugget/phoebot/lib/player"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type Mailbox struct {
	ID       uuid.UUID
	Added    time.Time
	Changed  time.Time
	LastScan time.Time
	Enabled  bool
	Class    string
	Owner    string
	Signtext string
	Material string
	World    string
	X        int
	Y        int
	Z        int
}

func MailboxFields() string {
	return "mailboxid, added, changed, lastscan, enabled, class, owner, signtext, material, world, x, y, z"
}

func (b *Mailbox) RowScan(rows *sql.Rows) error {
	return rows.Scan(
		&b.ID,
		&b.Added,
		&b.Changed,
		&b.LastScan,
		&b.Enabled,
		&b.Class,
		&b.Owner,
		&b.Signtext,
		&b.Material,
		&b.World,
		&b.X,
		&b.Y,
		&b.Z,
	)
}

func (b *Mailbox) Rename() error {

	fmt.Printf("Rename! %+v\n", b)

	customName := b.Owner + `\'s Mailbox`
	command := fmt.Sprintf(`/data modify block %d %d %d CustomName set value '{"text":"%s"}'`, b.X, b.Y, b.Z, customName)

	if ipc.ServerSayStream != nil {
		ipc.ServerSayStream <- command
	}

	return nil
}

func (b *Mailbox) Delete() error {
	query := `UPDATE mailbox SET deleted = current_timestamp at time zone 'utc' WHERE mailboxID = $1`
	_, err := db.DB.Exec(query, b.ID)
	return err
}

func NewMailbox(b Mailbox) (Mailbox, error) {
	existingBox, err := GetMailboxByOwnerAndLocation(b.Owner, b.World, b.X, b.Y, b.Z)
	if err != nil {
		return Mailbox{}, err
	}

	if existingBox.ID != uuid.Nil {
		fmt.Printf("existingBox! %+v\n", existingBox)
		return existingBox, nil
	}

	query := `INSERT INTO mailbox (class, owner, signtext, material, world, x, y, z)
			  SELECT $1, $2, $3, $4, $5, $6, $7, $8 RETURNING ` + MailboxFields()

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
	// Check for destruction
	fmt.Printf("Poll Mailbox: %+v\n", l)
	b, err := coreprotect.GetBlock(1, l.X, l.Y, l.Z)
	if err != nil {
		return err
	}

	if b.Action == "destroyed" {
		err := l.Delete()
		if err != nil {
			logrus.WithError(err).Error("Mailbox was destroyed, unable to remove record")
		} else {
			logrus.Warn("Mailbox was destroyed, removed record")

			message := fmt.Sprintf("Your mailbox at %d %d %d seems to be gone, so I will stop checking it.", l.X, l.Y, l.Z)
			err = player.SendMessage(l.Owner, message)
			if err != nil {
				logrus.WithError(err).Error("Unable to send message to player")
			}
		}

		return nil
	}

	fmt.Printf("Poll Mailboxesd:  %+v\n", b)

	return nil
}

func GetMailboxes() (ll []Mailbox, err error) {
	query := `SELECT ` + MailboxFields() + ` from mailbox WHERE deleted IS NULL AND enabled IS TRUE`

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

		ll = append(ll, b)
	}

	return ll, nil
}
