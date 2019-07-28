package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/nugget/phoebot/lib/db"
	"github.com/nugget/phoebot/lib/mapping"
	"github.com/nugget/phoebot/lib/phoelib"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func setupConfig() *viper.Viper {
	c := viper.New()
	c.AutomaticEnv()

	return c
}

func main() {
	config := setupConfig()
	dbURI := config.GetString("DATABASE_URI")
	debugLevel := config.GetString("PHOEBOT_DEBUG")

	if debugLevel != "" {
		_, err := phoelib.LogLevel(debugLevel)
		if err != nil {
			logrus.WithError(err).Error("Unable to set LogLevel")
		}
	}

	err := db.Connect(dbURI)
	if err != nil {
		logrus.WithError(err).Fatal("Unable to connect to database")
	}

	var (
		x        int
		z        int
		scale    int
		mapid    int
		coverage bool
	)

	flag.IntVar(&x, "x", 0, "X Position")
	flag.IntVar(&z, "z", 0, "X Position")
	flag.IntVar(&scale, "scale", 1, "Map Scale")
	flag.IntVar(&mapid, "id", 0, "Map ID")
	flag.BoolVar(&coverage, "coverage", false, "coverage for matching maps")
	flag.Parse()

	if coverage {
		mapList, err := mapping.GetByPosition(x, z)
		if err != nil {
			logrus.WithError(err).Fatal("GetByPosition Failure")
		}
		for _, m := range mapList {
			logrus.WithFields(logrus.Fields{
				"x":      x,
				"z":      z,
				"mapID":  m.MapID,
				"scale":  m.Scale,
				"leftX":  m.LeftX,
				"leftZ":  m.LeftZ,
				"rightX": m.RightX,
				"rightZ": m.RightZ,
			}).Info("Position covered by map")
		}
		os.Exit(0)
	}

	m := mapping.NewMap()

	m.Scale = scale
	m.MapID = mapid
	m.LeftX, m.LeftZ, m.RightX, m.RightZ = mapping.MapBoundaries(x, z, scale)

	logrus.Info(fmt.Sprintf("Map scaled 1:%d spans from (%d, %d) to (%d, %d)",
		m.Scale,
		m.LeftX, m.LeftZ,
		m.RightX, m.RightZ,
	))

	err = mapping.Update(m)
	if err != nil {
		logrus.WithError(err).Fatal("Unable to save map")
	}
}
