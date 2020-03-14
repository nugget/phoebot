package player

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/nugget/phoebot/lib/config"
	"github.com/nugget/phoebot/lib/db"
	"github.com/nugget/phoebot/lib/player"

	"github.com/sirupsen/logrus"
)

func OnJoinVerifyNag(joinMessage string) error {
	re := regexp.MustCompile(`^([^ ]+) (joined|left) the game`)
	res := re.FindStringSubmatch(joinMessage)

	if len(res) != 3 {
		// Unrecognized/Unparsable join message
		logrus.WithFields(logrus.Fields{
			"message": joinMessage,
			"res":     strings.Join(res, ":"),
			"len":     len(res),
		}).Warn("Unable to parse joinMessage")
		return nil
	}

	minecraftName := res[1]
	action := res[2]

	if action == "left" {
		// Ignore departures
		return nil
	}

	me, err := config.GetString("minecraftName", minecraftName)
	if err != nil {
		return err
	}
	if me == res[1] {
		// Ignore my own joins
		return nil
	}

	p, err := GetPlayerFromMinecraftName(minecraftName)
	if err != nil {
		return err
	}

	if p.Verified {
		logrus.WithFields(logrus.Fields{
			"player": p.MinecraftName,
		}).Trace("Player is already verified with Discord")

		err = player.Advancement(minecraftName, "phoenixcraft:phoenixcraft/discord")
		if err != nil {
			logrus.WithError(err).Warn("Unable to grant advancement")
		}

		return nil
	}

	code, lastNag, err := GenerateCode(minecraftName)
	if err != nil {
		return err
	}

	compareDate := lastNag.Add(time.Hour * 24 * 7)

	if lastNag.IsZero() || compareDate.Before(time.Now()) {
		nagMessage := fmt.Sprintf(`It would be great if I knew your Discord account name!  Hop on Discord and send me a private message that says: !verify %s`, code)
		p.SendMessage(nagMessage)
		logrus.WithFields(logrus.Fields{
			"player":      p.MinecraftName,
			"lastNag":     lastNag,
			"compareDate": compareDate,
			"now":         time.Now(),
		}).Info("Sent verification nag")
	} else {
		logrus.WithFields(logrus.Fields{
			"player":      p.MinecraftName,
			"lastNag":     lastNag,
			"compareDate": compareDate,
			"now":         time.Now(),
		}).Trace("Skipped verification nag")
	}

	return nil
}

func GenerateCode(minecraftName string) (string, time.Time, error) {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("abcdefghijkmnopqrstuvwxyz23456789")
	length := 5
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	code := strings.ToLower(b.String())

	query := `INSERT INTO verify (minecraftname, code)
			  SELECT $1, $2
			  ON CONFLICT (minecraftname) 
			  DO UPDATE SET code = $2, deleted = NULL
			  RETURNING lastnag`

	rows, err := db.DB.Query(query, minecraftName, code)
	if err != nil {
		return "", time.Now(), err
	}
	defer rows.Close()

	var lastNag time.Time

	for rows.Next() {
		err := rows.Scan(&lastNag)
		if err != nil {
			return "", time.Now(), err
		}
		return code, lastNag, nil
	}

	return "", time.Now(), fmt.Errorf("GenerateCode Error")
}

func LookupCode(code string) (string, time.Time, error) {
	query := `SELECT minecraftname, lastnag FROM verify WHERE deleted IS NULL AND code ILIKE $1`

	rows, err := db.DB.Query(query, code)
	if err != nil {
		return "", time.Now(), err
	}
	defer rows.Close()

	var (
		minecraftName string
		lastNag       time.Time
	)

	for rows.Next() {
		err := rows.Scan(
			&minecraftName,
			&lastNag,
		)
		if err != nil {
			return "", time.Now(), err
		}
		return minecraftName, lastNag, nil
	}

	return "", time.Now(), fmt.Errorf("LookupCode Error")

}
