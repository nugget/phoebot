package main

import (
	"os"

	"github.com/nugget/phoebot/lib/coreprotect"
)

func main() {
	coreprotect.Connect(os.Getenv("COREPROTECT_URI"))
	coreprotect.ScanBoxes()

}
