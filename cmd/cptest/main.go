package main

import (
	"os"

	"github.com/nugget/phoebot/lib/coreprotect"
	"github.com/nugget/phoebot/lib/db"
	"github.com/nugget/phoebot/lib/postal"
	"github.com/sirupsen/logrus"
)

func main() {
	var err error

	err = db.Connect(os.Getenv("DATABASE_URI"))
	if err != nil {
		logrus.WithError(err).Fatal("Unable to connect to database")
	}

	err = coreprotect.Connect(os.Getenv("COREPROTECT_URI"))
	if err != nil {
		logrus.WithError(err).Fatal("coreprotect.Connect Failed")
	}

	//	err = SearchForMailboxes(-35, 71, 152, -29, 69, 152)
	//
	// ScanBoxes(dimension string, lastScan time.Time, sx, sy, sz, fx, fy, fz int) error {

	err = postal.SearchServer()
	if err != nil {
		logrus.WithError(err).Fatal("SearchServer Failed")
	}

}
