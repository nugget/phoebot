package phoelib

import (
	"crypto/sha256"
	"fmt"
	"hash"
	"strings"

	"github.com/nugget/phoebot/lib/db"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

type Ignore struct {
	Category string
	Target   string
}

var (
	Ignores     []Ignore
	ignoresHash hash.Hash
)

func DebugSlice(res []string) {
	logrus.WithField("elements", len(res)).Debug("Dumper contents of 'res' slice:")
	for i, v := range res {
		logrus.Debugf("  %d: '%s' (%d)", i, v, len(v))
	}

}

func LogLevel(reqLevel string) (setLevel logrus.Level, err error) {
	reqLevel = strings.ToLower(reqLevel)

	switch reqLevel {
	case "trace":
		setLevel = logrus.TraceLevel
	case "debug":
		setLevel = logrus.DebugLevel
	case "info":
		setLevel = logrus.InfoLevel
	case "warn":
		setLevel = logrus.WarnLevel
	case "error":
		setLevel = logrus.ErrorLevel
	default:
		return setLevel, fmt.Errorf("Unrecognized log level '%s'", reqLevel)
	}

	logrus.SetLevel(setLevel)

	return setLevel, nil
}

func LogSQL(query string, args ...string) {
	query = strings.ReplaceAll(query, "\n", " ")
	query = strings.ReplaceAll(query, "\t", " ")

	logrus.WithFields(logrus.Fields{
		"query": query,
		"args":  args,
	}).Trace("Executing SQL Query")
}

func LoadIgnores() error {
	var (
		newIgnores []Ignore
		idList     []string
	)

	query := `SELECT ignoreid, category, target FROM ignore WHERE deleted IS NULL AND enabled IS TRUE ORDER BY added`

	rows, err := db.DB.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		i := Ignore{}
		id := ""

		err := rows.Scan(&id, &i.Category, &i.Target)
		if err != nil {
			return err
		}

		i.Category = strings.ToLower(i.Category)

		newIgnores = append(newIgnores, i)
		idList = append(idList, id)
	}
	logrus.WithField("idList", idList).Trace("idList Debug")

	newHash := sha256.New()
	newHash.Write([]byte(strings.Join(idList, ":")))

	if newHash == ignoresHash {
		logrus.WithFields(logrus.Fields{
			"count":    len(newIgnores),
			"listHash": fmt.Sprintf("%x", newHash.Sum(nil)),
		}).Trace("Ignores list unchanged")
	} else {
		logrus.WithFields(logrus.Fields{
			"count":    len(newIgnores),
			"listHash": fmt.Sprintf("%x", newHash.Sum(nil)),
		}).Info("Updated ignores list from database")

		Ignores = newIgnores
	}

	return nil
}

func IgnoreMessage(dm *discordgo.MessageCreate) bool {
	for _, i := range Ignores {
		switch i.Category {
		case "guildid":
			if i.Target == dm.GuildID {
				return true
			}
		case "playerid":
			if i.Target == dm.Author.ID {
				return true
			}
		case "channelid":
			if i.Target == dm.ChannelID {
				return true
			}
		default:
			logrus.WithField("category", i.Category).Warn("Unrecognized ignore category")
		}

	}

	return false
}

func PlayerHasACL(playerID, acl string) bool {
	query := `SELECT key FROM acl WHERE playerid = $1 AND key ILIKE $2 AND deleted IS NULL`

	LogSQL(query, playerID, acl)
	rows, err := db.DB.Query(query, playerID, acl)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":    err,
			"playerID": playerID,
			"acl":      acl,
		}).Error("SQL Error in PlayerHasACL")
		return false
	}
	defer rows.Close()

	for rows.Next() {
		return true
	}

	return false
}
