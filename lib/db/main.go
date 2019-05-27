package db

import (
	"database/sql"
	"fmt"

	"github.com/lib/pq"
	"github.com/nugget/phoebot/lib/builddata"
)

var (
	DB *sql.DB
)

func Connect(URIstring string) error {
	connString, err := pq.ParseURL(URIstring)
	if err != nil {
		return err
	}

	appName := fmt.Sprintf("Phoebot %s", builddata.VERSION)

	connString = fmt.Sprintf("%s application_name='%s'", connString, appName)

	DB, err = sql.Open("postgres", connString)
	return err
}
