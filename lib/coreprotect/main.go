package coreprotect

import (
	"database/sql"
	"fmt"
	"net/url"

	"github.com/nugget/phoebot/lib/phoelib"

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
	phoelib.LogSQL(query, id)
	row := DB.QueryRow(query, id)
	err := row.Scan(&world)
	if err != nil {
		logrus.WithError(err).Error("WorldFromWid error")
		return ""
	}

	World[id] = world

	logrus.WithFields(logrus.Fields{
		"wid":   World[id],
		"world": world,
	}).Trace("WorldFromWid")

	return World[id]
}

func WidFromWorld(world string) (wid int) {
	for k, v := range World {
		if v == world {
			return k
		}
	}

	query := `SELECT id FROM co_world WHERE world = ? LIMIT 1`
	phoelib.LogSQL(query, world)
	row := DB.QueryRow(query, world)
	err := row.Scan(&wid)
	if err != nil {
		logrus.WithError(err).Error("WidFromWorld error")
		return 0
	}

	World[wid] = world

	logrus.WithFields(logrus.Fields{
		"wid":   wid,
		"world": World[wid],
	}).Trace("WidFromWorld")

	return wid
}
