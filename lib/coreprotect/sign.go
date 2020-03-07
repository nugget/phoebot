package coreprotect

import (
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/nugget/phoebot/lib/phoelib"
	"github.com/sirupsen/logrus"
)

type SignLog struct {
	Epoch     int64
	Timestamp time.Time
	User      string
	UserID    int
	World     string
	WorldID   int
	X         int
	Y         int
	Z         int
	Color     int
	Line1     string
	Line2     string
	Line3     string
	Line4     string
}

func (s *SignLog) Parse() error {
	s.Timestamp = time.Unix(s.Epoch, 0)

	return nil
}

func ScanSigns(matchString string, lastScan time.Time) (l []SignLog, err error) {
	epoch := lastScan.Unix()

	matchString = "%" + matchString + "%" // SQL wildcard syntax

	query := `SELECT s.time, s.user as userid, u.user, s.wid, w.world, s.x, s.y, s.z, s.color, s.line_1, s.line_2, s.line_3, s.line_4
	          FROM co_sign s
			  LEFT JOIN (co_user u, co_world w) on (s.user = u.rowid AND w.rowid = s.wid)
			  WHERE s.time > ?
			  AND (
				  line_1 LIKE ? OR
				  line_2 LIKE ? OR
				  line_3 LIKE ? OR
				  line_4 LIKE ? 
				  )
			  ORDER BY s.time`

	logrus.WithFields(logrus.Fields{
		"lastScan": lastScan,
		"epoch":    epoch,
		"match":    matchString,
	}).Trace("Looking for sign activity")

	phoelib.LogSQL(query, lastScan, matchString)

	rows, err := DB.Query(query, lastScan, matchString, matchString, matchString, matchString)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		s := SignLog{}

		err = rows.Scan(
			&s.Epoch,
			&s.UserID,
			&s.User,
			&s.WorldID,
			&s.World,
			&s.X,
			&s.Y,
			&s.Z,
			&s.Color,
			&s.Line1,
			&s.Line2,
			&s.Line3,
			&s.Line4,
		)
		if err != nil {
			return nil, err
		}
		s.Parse()

		l = append(l, s)
	}

	return l, nil
}
