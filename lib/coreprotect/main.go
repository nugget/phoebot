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
