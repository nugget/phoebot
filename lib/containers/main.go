package containers

import (
	"fmt"
	"time"

	"github.com/nugget/phoebot/lib/db"
	"github.com/sirupsen/logrus"
)

type ScanRange struct {
	ScanRangeID string
	LastScan    time.Time
	CurrentTime time.Time
	Name        string
	Owner       string
	Dimension   string
	Sx          int
	Sy          int
	Sz          int
	Fx          int
	Fy          int
	Fz          int
}

func GetRanges(scanType string) (ranges []ScanRange, err error) {
	query := `SELECT scanrangeID, lastscan, current_timestamp, name, owner, dimension, sx, sy, sz, fx, fy, fz FROM scanrange WHERE enabled IS TRUE AND deleted IS NULL AND scantype = $1`

	rows, err := db.DB.Query(query, scanType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		sr := ScanRange{}

		err := rows.Scan(
			&sr.ScanRangeID,
			&sr.LastScan,
			&sr.CurrentTime,
			&sr.Name,
			&sr.Owner,
			&sr.Dimension,
			&sr.Sx, &sr.Sy, &sr.Sz,
			&sr.Fx, &sr.Fy, &sr.Fz,
		)
		if err != nil {
			return nil, err
		}

		logrus.WithFields(logrus.Fields{
			"scanRangeID": sr.ScanRangeID,
			"currentTime": sr.CurrentTime,
			"lastScan":    sr.LastScan,
			"dimension":   sr.Dimension,
			"name":        sr.Name,
			"start":       fmt.Sprintf("(%d, %d, %d)", sr.Sx, sr.Sy, sr.Sz),
			"finish":      fmt.Sprintf("(%d, %d, %d)", sr.Fx, sr.Fy, sr.Fz),
		}).Trace("ScanRange loaded from database")

		query := `UPDATE scanrange SET lastScan = $1 WHERE scanrangeID = $2`
		_, err = db.DB.Exec(query, sr.CurrentTime, sr.ScanRangeID)

		if err != nil {
			logrus.WithFields(logrus.Fields{
				"id":       sr.ScanRangeID,
				"lastScan": sr.CurrentTime,
				"err":      err,
			}).Error("Unable to update last scan time")
		}

		ranges = append(ranges, sr)

	}

	return ranges, nil
}
