package phoelib

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

func DebugSlice(res []string) {
	logrus.WithField("elements", len(res)).Debug("Dumper contents of 'res' slice:")
	for i, v := range res {
		logrus.Debugf("  %d: '%s' (%d)", i, v, len(v))
	}

}

func LogLevel(reqLevel string) (setLevel logrus.Level, err error) {
	reqLevel = strings.ToLower(reqLevel)

	switch reqLevel {
	case "trace":
		setLevel = logrus.TraceLevel
	case "debug":
		setLevel = logrus.DebugLevel
	case "info":
		setLevel = logrus.InfoLevel
	case "warn":
		setLevel = logrus.WarnLevel
	case "error":
		setLevel = logrus.ErrorLevel
	default:
		return setLevel, fmt.Errorf("Unrecognized log level '%s'", reqLevel)
	}

	logrus.SetLevel(setLevel)

	return setLevel, nil
}

func LogSQL(query string, args ...string) {
	query = strings.ReplaceAll(query, "\n", " ")
	query = strings.ReplaceAll(query, "\t", " ")

	logrus.WithFields(logrus.Fields{
		"query": query,
		"args":  args,
	}).Debug("Executing SQL Query")
}
