package console

import (
	"fmt"
	"regexp"

	"github.com/sirupsen/logrus"
)

func GetBlock(x, y, z int, path string) (string, error) {
	command := fmt.Sprintf("data get block %v %v %v %v", x, y, z, path)
	data, err := s.sendCommand(command)
	if err != nil {
		return "", err
	}

	logrus.WithFields(logrus.Fields{
		"x":      x,
		"y":      y,
		"z":      z,
		"path":   path,
		"result": data,
		"err":    err,
	}).Trace("console.GetBlock")

	r := regexp.MustCompile(`: ({.*})`)
	res := r.FindStringSubmatch(data)

	if len(res) != 2 {
		return "", fmt.Errorf("Unable to parse block data")
	}

	if res[1] == "" {
		return "", fmt.Errorf("Empty block data")
	}

	return res[1], nil
}
