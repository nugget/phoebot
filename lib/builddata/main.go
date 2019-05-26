package builddata

// These values are populated by Bazel during the build process.  There is
// no longer a required `go generate` step to populate these values.  If you
// build by hand, outside of the Bazel workspace, these values will remain as
// the defaults defined below.
//
// See the BUILD.bazel x_defs for the list and mappings.
//

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

var (
	VERSION         string = "unknown"
	GITBRANCH       string = "unknown"
	GITCOMMIT       string = "unknown"
	GOVERSION       string = "unknown"
	BUILDENV        string = "unknown"
	BUILDEPOCH_STR  string = "0"
	BUILDEPOCH      int64  = 0
	BUILDDATE       string = "1970-01-01 00:00 UTC"
	BUILDHOST       string = "unknown"
	BUILDUSER       string = "unknown"
	BUILDEMBEDLABEL string = "unknown"
)

func LogStructured() {
	logrus.WithFields(logrus.Fields{
		"version":   VERSION,
		"gitBranch": GITBRANCH,
		"gitCommit": GITCOMMIT,
		"buildTime": BUILDDATE,
		"buildHost": BUILDHOST,
		"goVersion": GOVERSION,
		"user":      BUILDUSER,
		"host":      BUILDHOST,
		"os":        BUILDENV,
	}).Info("Build Data")
}

func LogConversational() {
	logrus.Info(Version())
	logrus.Info(BuiltBy())
	logrus.Info(GitInfo())
	logrus.Info(EmbedLabel())
}

func Version() string {
	return fmt.Sprintf("Phoebot/%s (%s)",
		VERSION,
		BUILDDATE,
	)
}

func BuiltBy() string {
	return fmt.Sprintf("Built by %s@%s running %s",
		BUILDUSER,
		BUILDHOST,
		BUILDENV,
	)
}

func GitInfo() string {
	return fmt.Sprintf("Branch `%s` commit `%s`\n",
		GITBRANCH,
		GITCOMMIT,
	)
}

func EmbedLabel() string {
	if BUILDEMBEDLABEL != "" {
		return fmt.Sprintf("Label: %s\n", BUILDEMBEDLABEL)
	}

	return ""
}

func Uname() string {
	buf := strings.Builder{}

	buf.WriteString(Version() + "\n")
	buf.WriteString(BuiltBy() + "\n")
	buf.WriteString(GitInfo() + "\n")
	buf.WriteString("Source code and issue tracker are at https://github.com/nugget/phoebot\n")

	return buf.String()
}
