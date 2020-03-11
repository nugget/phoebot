package coreprotect

import (
	"database/sql"
	"fmt"
	"net/url"

	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
)

var (
	DB    *sql.DB
	World map[int]string
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

	if World == nil {
		World = make(map[int]string)
	}

	return nil
}

func WorldFromWid(id int) (world string) {
	if World[id] != "" {
		return World[id]
	}

	query := `SELECT world FROM co_world WHERE id = ? LIMIT 1`
	row := DB.QueryRow(query)
	err := row.Scan(&world)
	if err != nil {
		return ""
	}

	World[id] = world
	return World[id]
}

func WidFromWorld(world string) (wid int) {
	for k, v := range World {
		if v == world {
			return k
		}
	}

	query := `SELECT id FROM co_world WHERE world = ? LIMIT 1`
	row := DB.QueryRow(query)
	err := row.Scan(&wid)
	if err != nil {
		return 0
	}

	World[wid] = world
	return wid
}
