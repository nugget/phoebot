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
	query := `SELECT value FROM config WHERE key = $1`

	phoelib.LogSQL(query, item)

	rows, err := db.DB.Query(query, item)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"key":   item,
		}).Error("SQL Error in GetString")
		return "", err
	}
	defer rows.Close()

	for rows.Next() {
		var value string
		err = rows.Scan(&value)
		return value, nil
	}

	return defaultValue, nil
}

func WriteString(item string, value string) error {
	query := `INSERT INTO config (key, value) VALUES ($1, $2)
	          ON CONFLICT (key) DO UPDATE
			      SET value = $2`

	phoelib.LogSQL(query, item, value)

	logrus.WithFields(logrus.Fields{
		"item":  item,
		"value": value,
	}).Trace("config.WriteString")

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

	tT := time.Unix(epoch, 0)

	logrus.WithFields(logrus.Fields{
		"valueString": valueString,
		"epoch":       epoch,
		"time":        tT,
	}).Trace("GetTime processing")

	return tT, nil
}

func WriteTime(item string, value time.Time) error {
	valueString := fmt.Sprintf("%d", value.Unix())

	return WriteString(item, valueString)
}
