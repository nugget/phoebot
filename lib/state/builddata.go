package state

// These values are populated by Bazel during the build process.  There is
// no longer a required `go generate` step to populate these values.  If you
// build by hand, outside of the Bazel workspace, these values will remain as
// the defaults defined below.
//
// See the BUILD.bazel x_defs for the list and mappings.
//

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
