package subscriptions

import (
	"fmt"
	"strings"

	"github.com/nugget/phoebot/lib/db"
	"github.com/nugget/phoebot/lib/discord"
	"github.com/nugget/phoebot/lib/phoelib"
	"github.com/nugget/phoebot/models"

	"github.com/sirupsen/logrus"
)

func SubscriptionsMatch(a, b models.Subscription) bool {
	if a.ChannelID == b.ChannelID {
		if strings.ToLower(a.Class) == strings.ToLower(b.Class) {
			if strings.ToLower(a.Name) == strings.ToLower(b.Name) {
				return true
			}
		}
	}

	return false
}

func GetMatching(class, name string) (sList []models.Subscription, err error) {
	query := `SELECT channelid, class, name, target FROM subscription
			  WHERE deleted IS NULL AND
			        class ILIKE $1 AND name ILIKE $2`

	phoelib.LogSQL(query, class, name)
	rows, err := db.DB.Query(query, class, name)
	if err != nil {
		return sList, err
	}
	defer rows.Close()

	for rows.Next() {
		sub := models.Subscription{}
		err := rows.Scan(
			&sub.ChannelID,
			&sub.Class,
			&sub.Name,
			&sub.Target,
		)
		if err != nil {
			return sList, err
		}

		sList = append(sList, sub)
	}

	return sList, err
}

func GetByChannel(channelid string) (sList []models.Subscription, err error) {
	query := `SELECT channelid, class, name, target FROM subscription
			  WHERE deleted IS NULL AND channelid ILIKE $1`

	rows, err := db.DB.Query(query, channelid)
	if err != nil {
		return sList, err
	}
	defer rows.Close()

	for rows.Next() {
		sub := models.Subscription{}
		err := rows.Scan(
			&sub.ChannelID,
			&sub.Class,
			&sub.Name,
			&sub.Target,
		)
		if err != nil {
			return sList, err
		}

		sList = append(sList, sub)
	}

	return sList, err
}

func SubscriptionExists(sub models.Subscription) (bool, error) {
	logrus.WithFields(logrus.Fields{
		"sub": sub,
	}).Debug("SubscriptionExists")

	query := `SELECT count(*) AS exists 
              FROM subscription
			  WHERE 
			    deleted IS NULL AND
			    channelID = $1 AND class = $2 AND name = $3`

	rows, err := db.DB.Query(query, sub.ChannelID, sub.Class, sub.Name)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"query": query,
		}).Error("SQL Error")

		return false, err
	}
	defer rows.Close()

	var count int

	for rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			logrus.WithError(err).Error("rows.Scan error")
			return false, err
		}

		if count > 0 {
			return true, nil
		}
	}

	return false, nil
}

func AddSubscription(sub models.Subscription) error {
	if sub.Class == "" || sub.Name == "" {
		return fmt.Errorf("'%s' '%s' doesn't look like a real thing to me.", sub.Class, sub.Name)
	} else {
		exists, err := SubscriptionExists(sub)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("A subscription for %s/%s already exists for this channel", sub.Class, sub.Name)
		} else {
			query := `INSERT INTO subscription (channelid, class, name, target) SELECT $1, $2, $3, $4`

			_, err := db.DB.Exec(query, sub.ChannelID, sub.Class, sub.Name, sub.Target)
			if err != nil {
				return err
			}

			message := fmt.Sprintf("You are now subscribed to receive updates to this channel for %s releases from %s", sub.Name, sub.Class)
			discord.Session.ChannelMessageSend(sub.ChannelID, message)

			logrus.WithFields(logrus.Fields{
				"name":      sub.Name,
				"class":     sub.Class,
				"channelID": sub.ChannelID,
				"target":    sub.Target,
			}).Info("Added new subscription")
		}
	}

	return nil
}

func DropSubscription(sub models.Subscription) error {
	query := `UPDATE subscription SET deleted = current_timestamp
	          WHERE deleted IS NULL AND channelid = $1 AND class = $2 AND name = $3
			  RETURNING *`

	rows, err := db.DB.Query(query, sub.ChannelID, sub.Class, sub.Name)
	if err != nil {
		return err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		logrus.WithField("row", rows).Debug("Dropped subscription")
		count++
	}

	if count > 0 {
		message := fmt.Sprintf("You are no longer subscribed to receive updates to this channel for %s releases from %s", sub.Name, sub.Class)
		discord.Session.ChannelMessageSend(sub.ChannelID, message)
	}

	return nil
}
