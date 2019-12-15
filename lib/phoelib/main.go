package phoelib

import (
	"crypto/sha256"
	"fmt"
	"hash"
	"os"
	"path/filepath"
	"strings"

	"github.com/nugget/phoebot/lib/builddata"
	"github.com/nugget/phoebot/lib/db"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"

	graylog "github.com/gemnasium/logrus-graylog-hook/v3"
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

func Graylog() error {
	appName := filepath.Base(os.Args[0])

	hook := graylog.NewGraylogHook("172.28.10.11:12201", map[string]interface{}{
		"service": appName,
		"version": builddata.Version(),
	})
	logrus.AddHook(hook)

	return nil
}

func LogSQL(query string, args ...interface{}) {
	query = strings.ReplaceAll(query, "\n", " ")
	query = strings.ReplaceAll(query, "\t", " ")
	query = strings.Join(strings.Fields(query), " ")

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
		}).Debug("Updated ignores list from database")

		Ignores = newIgnores
		ignoresHash = newHash
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

func Rebound(sx, sy, sz, fx, fy, fz int) (int, int, int, int, int, int) {
	if sx > fx {
		sx, fx = fx, sx
	}

	if sy > fy {
		sy, fy = fy, sy
	}

	if sz > fz {
		sz, fz = fz, sz
	}

	return sx, sy, sz, fx, fy, fz
}

func SizeOf(sx, sy, sz, fx, fy, fz int) int {
	sx, sy, sz, fx, fy, fz = Rebound(sx, sy, sz, fx, fy, fz)

	a := fx - sx + 1
	b := fy - sy + 1
	c := fz - sz + 1

	return a * b * c
}
