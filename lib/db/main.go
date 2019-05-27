package db

import (
	"database/sql"
	"fmt"
	"net/url"

	"github.com/lib/pq"
	"github.com/nugget/phoebot/lib/builddata"
	"github.com/sirupsen/logrus"
)

var (
	DB *sql.DB
)

func Connect(URIstring string) error {
	connString, err := pq.ParseURL(URIstring)
	if err != nil {
		return err
	}

	u, err := url.Parse(URIstring)
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"host":     u.Host,
		"user":     u.User.Username(),
		"database": u.Path,
	}).Info("Connecting to database")

	appName := fmt.Sprintf("Phoebot %s", builddata.VERSION)

	connString = fmt.Sprintf("%s application_name='%s'", connString, appName)

	DB, err = sql.Open("postgres", connString)
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

	logrus.WithField("version", version).Debug("Database version")

	return nil
}
