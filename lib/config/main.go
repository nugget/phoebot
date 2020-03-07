package config

import (
	"fmt"
	"strconv"
	"time"

	"github.com/nugget/phoebot/lib/db"
	"github.com/nugget/phoebot/lib/phoelib"

	"github.com/sirupsen/logrus"
)

func GetString(item string, defaultValue string) (string, error) {
	query := `SELECT key, value FROM config WHERE key = $1`

	phoelib.LogSQL(query, item)
	rows, err := db.DB.Query(query, item)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"key":   item,
		}).Error("SQL Error in PlayerHasACL")
		return "", err
	}
	defer rows.Close()

	for rows.Next() {
		var value string

		err = rows.Scan(
			&value,
		)
		return value, nil
	}

	return defaultValue, nil
}

func WriteString(item string, value string) error {
	query := `INSERT INTO config (key, value) VALUES ($1, $2)
	          ON CONFLICT (key) DO UPDATE
			      SET value = $2 WHERE key = $1`
	phoelib.LogSQL(query, item, value)
	_, err := db.DB.Exec(query, item, value)
	return err
}

func GetTime(item string, defaultValue time.Time) (t time.Time, err error) {
	valueString, err := GetString(item, "1")
	if err != nil {
		return t, err
	}

	epoch, err := strconv.ParseInt(valueString, 10, 64)
	if err != nil {
		return t, err
	}

	return time.Unix(epoch, 0), nil
}

func WriteTime(item string, value time.Time) error {
	valueString := fmt.Sprintf("%d", value.Unix())

	return WriteString(item, valueString)
}
