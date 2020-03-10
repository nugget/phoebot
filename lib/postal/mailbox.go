package postal

import (
	"database/sql"
	"time"

	"github.com/nugget/phoebot/lib/db"
	"github.com/nugget/phoebot/lib/phoelib"

	"github.com/google/uuid"
)

type Mailbox struct {
	ID       uuid.UUID
	Added    time.Time
	Changed  time.Time
	Deleted  time.Time
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

func (b *Mailbox) Fields() string {
	return "mailboxid, added, changed, deleted, lastscan, enabled, class, owner, signtext, material, world, x, y, z"
}

func (b *Mailbox) RowScan(rows *sql.Rows) error {
	rows.Scan(
		&b.ID,
		&b.Added,
		&b.Changed,
		&b.Deleted,
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
	return nil
}

func NewMailbox(b Mailbox) (Mailbox, error) {
	existingBox, err := GetMailboxByOwnerAndLocation(b.Owner, b.World, b.X, b.Y, b.Z)
	if err != nil {
		return Mailbox{}, err
	}

	if existingBox.ID != uuid.Nil {
		return existingBox, nil
	}

	query := `INSERT INTO mailbox (class, owner, signtext, material, world, x, y, z)
			  SELECT $1, $2, $3, $4, $5, $6, $7, $8 RETURNING ` + b.Fields()

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
	b := Mailbox{}

	query := `SELECT ` + b.Fields() + ` FROM mailbox WHERE deleted IS NULL AND owner = $1 AND world = $2
	AND x = $3 AND y = $4 AND z = $5`

	phoelib.LogSQL(query, owner, world, x, y, z)
	rows, err := db.DB.Query(query, owner, world, x, y, z)

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
