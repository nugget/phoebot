package console

import (
	"fmt"
	"math"
	"regexp"
	"strconv"

	"github.com/sirupsen/logrus"
)

type Player struct {
	Name      string
	X         int
	Y         int
	Z         int
	Dimension string
}

type ParseFunction func(string) error

func (p *Player) ParsePos(data string) error {
	p.X = 0
	p.Y = 0
	p.Z = 0

	r := regexp.MustCompile(`: \[([^d]+)d, ([^d]+)d, ([^d]+)d\]`)
	res := r.FindStringSubmatch(data)

	if len(res) != 4 {
		return fmt.Errorf("Unable to parse Pos data")
	}

	xf, err := strconv.ParseFloat(res[1], 64)
	if err != nil {
		return err
	}
	yf, err := strconv.ParseFloat(res[2], 64)
	if err != nil {
		return err
	}
	zf, err := strconv.ParseFloat(res[3], 64)
	if err != nil {
		return err
	}

	p.X = int(math.Floor(xf))
	p.Y = int(math.Floor(yf))
	p.Z = int(math.Floor(zf))

	return nil
}

func (p *Player) ParseDimension(data string) error {
	r := regexp.MustCompile(`: (.+)`)
	res := r.FindStringSubmatch(data)

	if len(res) != 2 {
		return fmt.Errorf("Unable to parse Dimension data")
	}

	switch res[1] {
	case "-1":
		p.Dimension = "nether"
	case "0":
		p.Dimension = "overworld"
	case "1":
		p.Dimension = "end"
	default:
		p.Dimension = "unknown"
	}

	return nil
}

func (p *Player) GetData(path string, fn ParseFunction) error {
	command := fmt.Sprintf("data get entity %s %s", p.Name, path)

	result, err := s.sendCommand(command)

	logrus.WithFields(logrus.Fields{
		"path":    path,
		"fn":      fn,
		"command": command,
		"result":  result,
		"err":     err,
	}).Trace("console.GetData")

	if err != nil {
		return err
	}
	err = fn(result)
	if err != nil {
		return err
	}

	return nil
}

func GetPlayer(name string) (Player, error) {
	p := Player{}
	p.Name = name

	p.GetData("Pos", p.ParsePos)
	p.GetData("Dimension", p.ParseDimension)

	return p, nil
}
