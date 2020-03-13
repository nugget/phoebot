package main

import (
	"os"

	"github.com/nugget/phoebot/lib/coreprotect"
	"github.com/nugget/phoebot/lib/db"
	"github.com/nugget/phoebot/lib/postal"
	"github.com/sirupsen/logrus"
)

// D1: -146 65 143
// D3: -146 64 143

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

	// Look for new tagging signs
	err = postal.NewSignScan()
	if err != nil {
		logrus.WithError(err).Error("postal.NewSignScan failed")
	}

	// Look for mailbox updates
	err = postal.PollMailboxes()
	if err != nil {
		logrus.WithError(err).Error("postal.PollMailboxes failed")
	}

}
