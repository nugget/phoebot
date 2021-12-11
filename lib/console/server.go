package console

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

type ServerInfo struct {
	Version    string
	Tps1       float64
	Tps5       float64
	Tps15      float64
	MaxPlayers int
	Players    []string
}

func tpsfloat(val string) (float64, error) {
	r := regexp.MustCompile(`([0-9.]+)`)
	res := r.FindStringSubmatch(val)
	if len(res) != 2 {
		return 0.0, fmt.Errorf("Can't clean val: '%s'", val)
	}
	return strconv.ParseFloat(res[1], 64)
}

func (si *ServerInfo) GetTPS() error {
	result, err := s.sendCommand("tps")
	if err != nil {
		return err
	}

	r := regexp.MustCompile(`: ([^,]+), ([^,]+), ([^,]+)`)
	res := r.FindStringSubmatch(result)
	if err != nil {
		return err
	}

	if len(res) != 4 {
		return fmt.Errorf("Unexpected tps result: '%v'", res)
	}

	si.Tps1, err = tpsfloat(res[1])
	if err != nil {
		return err
	}
	si.Tps5, err = tpsfloat(res[2])
	if err != nil {
		return err
	}
	si.Tps15, err = tpsfloat(res[3])
	if err != nil {
		return err
	}

	return nil
}

func (si *ServerInfo) GetVersion() error {
	result, err := s.sendCommand("version")
	if err != nil {
		return err
	}

	r := regexp.MustCompile(`This server is running (.*)`)
	res := r.FindStringSubmatch(result)
	if len(res) == 2 {
		si.Version = res[1]
	}

	return nil
}

func (si *ServerInfo) GetPlayers() error {
	si.MaxPlayers = 0
	si.Players = []string{}

	result, err := s.sendCommand("list")
	if err != nil {
		return err
	}

	rMax := regexp.MustCompile(`max ([0-9]+)`)
	res := rMax.FindStringSubmatch(result)
	if len(res) == 2 {
		si.MaxPlayers, _ = strconv.Atoi(res[1])
	}

	rList := regexp.MustCompile(`: (.*)`)
	res = rList.FindStringSubmatch(result)
	if len(res) == 2 {
		buf := strings.Split(res[1], ",")
		for _, p := range buf {
			si.Players = append(si.Players, strings.TrimSpace(p))
		}
	}

	return nil
}

func GetServerInfo() (ServerInfo, error) {
	si := ServerInfo{}

	err := si.GetTPS()
	if err != nil {
		logrus.WithError(err).Error("GetTPS failure")
	}

	err = si.GetVersion()
	if err != nil {
		logrus.WithError(err).Error("GetVersion failure")
	}

	err = si.GetPlayers()
	if err != nil {
		logrus.WithError(err).Error("GetPlayers failure")
	}

	return si, nil
}
