package coreprotect

import (
	"database/sql"
	"fmt"
	"net/url"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
)

var (
	DB *sql.DB
)

type containerLog struct {
	Epoch      int64
	Timestamp  time.Time
	User       string
	X          int
	Y          int
	Z          int
	Material   string
	Amount     int
	Action     int
	RolledBack int
}

func (c *containerLog) Parse() error {
	c.Timestamp = time.Unix(c.Epoch, 0)
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

func ScanBoxes() error {
	query := `SELECT
				c.time, u.user, x, y, z, m.material, c.amount, c.action, c.rolled_back
		   	  FROM co_container c 
			  LEFT JOIN (co_user u, co_material_map m) on (c.type = m.rowid and c.user = u.rowid)
			  WHERE x >= -35 and x <= -29 and y >= 69 and y <= 71 and z = 152
			  ORDER BY time DESC LIMIT 10`

	rows, err := DB.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		c := containerLog{}

		err = rows.Scan(
			&c.Epoch,
			&c.User,
			&c.X,
			&c.Y,
			&c.Z,
			&c.Material,
			&c.Amount,
			&c.Action,
			&c.RolledBack,
		)
		if err != nil {
			return err
		}
		c.Parse()

		fmt.Printf("%+v\n", c)
		fmt.Println("-- ")

	}

	return nil
}
