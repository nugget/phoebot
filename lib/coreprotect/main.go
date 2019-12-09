package coreprotect

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/nugget/phoebot/lib/phoelib"
	"github.com/sirupsen/logrus"
)

var (
	DB *sql.DB
)

type ContainerLog struct {
	Epoch       int64
	Timestamp   time.Time
	User        string
	X           int
	Y           int
	Z           int
	Material    string
	Amount      int
	Action      string
	ActionCode  int
	RolledBack  int
	Preposition string
}

func (c *ContainerLog) Parse() error {
	c.Timestamp = time.Unix(c.Epoch, 0)

	if c.ActionCode == 1 {
		c.Action = "placed"
		c.Preposition = "into"
	} else {
		c.Action = "took"
		c.Preposition = "from"
	}

	c.Material = strings.Title(strings.TrimPrefix(c.Material, "minecraft:"))

	return nil
}

func Connect(URIstring string) error {
	u, err := url.Parse(URIstring)
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"host":     u.Host,
		"user":     u.User.Username(),
		"database": u.Path,
	}).Info("Connecting to CoreProtect")

	p, _ := u.User.Password()
	connString := fmt.Sprintf("%s:%s@tcp(%s)%s", u.User.Username(), p, u.Host, u.Path)

	DB, err = sql.Open(u.Scheme, connString)
	if err != nil {
		return err
	}

	query := `SELECT version()`
	row := DB.QueryRow(query)

	version := "unknown"
	err = row.Scan(&version)
	if err != nil {
		return err
	}

	logrus.WithField("version", version).Debug("CoreProtect database version")

	return nil
}

func ScanContainers(dimension string, lastScan time.Time, sx, sy, sz, fx, fy, fz int) (l []ContainerLog, err error) {
	wid := 1
	epoch := lastScan.Unix()

	sx, sy, sz, fx, fy, fz = phoelib.Rebound(sx, sy, sz, fx, fy, fz)

	query := `SELECT
				u.user, x, y, z, m.material, c.action, c.rolled_back, max(c.time), sum(c.amount)
		   	  FROM co_container c 
			  LEFT JOIN (co_user u, co_material_map m) on (c.type = m.rowid and c.user = u.rowid)
			  WHERE c.rolled_back = 0 
			    AND c.wid = ?
			    AND c.x >= ? AND c.x <= ? 
				AND c.y >= ? AND c.y <= ? 
				AND c.z >= ? AND c.z <= ?
				AND c.time > ?
			  GROUP BY u.user, x, y, z, m.material, c.action, c.rolled_back
			  ORDER BY max(c.time)`

	logrus.WithFields(logrus.Fields{
		"lastScan":  lastScan,
		"epoch":     epoch,
		"dimension": dimension,
		"wid":       wid,
		"start":     fmt.Sprintf("(%d, %d, %d)", sx, sy, sz),
		"finish":    fmt.Sprintf("(%d, %d, %d)", fx, fy, fz),
	}).Trace("Looking for container activity")

	rows, err := DB.Query(query, wid, sx, fx, sy, fy, sz, fz, epoch)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		c := ContainerLog{}

		err = rows.Scan(
			&c.User,
			&c.X,
			&c.Y,
			&c.Z,
			&c.Material,
			&c.ActionCode,
			&c.RolledBack,
			&c.Epoch,
			&c.Amount,
		)
		if err != nil {
			return nil, err
		}
		c.Parse()

		l = append(l, c)
	}

	return l, nil
}
