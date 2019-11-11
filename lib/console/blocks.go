package console

import (
	"fmt"
	"regexp"
	"strings"

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

	r := regexp.MustCompile(`data: (.*)`)
	res := r.FindStringSubmatch(data)

	if len(res) != 2 {
		logrus.WithFields(logrus.Fields{
			"x":      x,
			"y":      y,
			"z":      z,
			"path":   path,
			"result": data,
		}).Debug("Bad block data")
		return "", fmt.Errorf("Unable to parse block data")
	}

	if res[1] == "" {
		return "", fmt.Errorf("Empty block data")
	}

	return res[1], nil
}

func GetString(x, y, z int, path string) (string, error) {
	data, err := GetBlock(x, y, z, path)
	if err != nil {
		return "", err
	}

	t := ""

	pattern := fmt.Sprintf(`"([^"]+)"`)
	re := regexp.MustCompile(pattern)
	res := re.FindStringSubmatch(data)
	if len(res) == 2 {
		t = res[1]
	}

	t = strings.ReplaceAll(t, `\'`, `'`)

	return t, nil

}

func GetText(x, y, z int, path string) (string, error) {
	data, err := GetBlock(x, y, z, path)
	if err != nil {
		return "", err
	}

	t := ""

	pattern := fmt.Sprintf(`'{"text":"([^"]+)"`)
	re := regexp.MustCompile(pattern)
	res := re.FindStringSubmatch(data)
	if len(res) == 2 {
		t = res[1]
	}

	t = strings.ReplaceAll(t, `\'`, `'`)

	return t, nil
}
