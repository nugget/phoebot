package main

import (
	"os"

	"github.com/nugget/phoebot/lib/coreprotect"
	"github.com/sirupsen/logrus"
)

func main() {
	var err error

	err = coreprotect.Connect(os.Getenv("COREPROTECT_URI"))
	if err != nil {
		logrus.WithError(err).Fatal("coreprotect.Connect Failed")
	}
	err = coreprotect.ScanBoxes()
	if err != nil {
		logrus.WithError(err).Fatal("coreprotect.ScanBoxes Failed")
	}

}
