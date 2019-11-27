package coreprotect

import (
	"database/sql"
	"fmt"
	"net/url"

	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
)

var (
	DB *sql.DB
)

type containerLog struct {
	RowID      int
	Epoch      int
	User       int
	Wid        int
	X          int
	Y          int
	Z          int
	Type       int
	Data       int
	Amount     int
	Metadata   []byte
	Action     int
	RolledBack int
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
	connString := fmt.Sprintf("%s:%s@tcp(%s)/%s", u.User, u.User.Username(), p, u.Host, u.Path)

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
	query := `SELECT rowid, time, user, wid, x, y, z, type,
                     data, amount, metadata, action, rolled_back
	          FROM co_container
			  LEFT JOIN co_material_map 
			  WHERE user = 28 AND type = 26 ORDER BY time`

	rows, err := DB.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		c := containerLog{}

		err = rows.Scan(
			&c.RowID,
			&c.Epoch,
			&c.User,
			&c.Wid,
			&c.X,
			&c.Y,
			&c.Z,
			&c.Type,
			&c.Data,
			&c.Amount,
			&c.Metadata,
			&c.Action,
			&c.RolledBack,
		)
		if err != nil {
			panic(err.Error())
		}

		fmt.Printf("%+v\n", c)
		fmt.Printf("%s\n", c.Metadata)
		fmt.Println("-- ")

	}

	return nil
}
