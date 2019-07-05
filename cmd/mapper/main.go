package main

import (
	"flag"
	"fmt"

	"github.com/nugget/phoebot/lib/db"
	"github.com/nugget/phoebot/lib/mapping"

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

	err := db.Connect(dbURI)
	if err != nil {
		logrus.WithError(err).Fatal("Unable to connect to database")
	}

	var (
		x     int
		z     int
		scale int
		mapid int
	)

	flag.IntVar(&x, "x", 0, "X Position")
	flag.IntVar(&z, "z", 0, "X Position")
	flag.IntVar(&scale, "scale", 1, "Map Scale")
	flag.IntVar(&mapid, "id", 0, "Map ID")
	flag.Parse()

	m := mapping.NewMap()

	m.Scale = scale
	m.MapID = mapid
	m.LeftX, m.LeftZ, m.RightX, m.RightZ = mapping.MapBoundaries(x, z, scale)

	fmt.Printf("Map scaled 1:%d spans from (%d, %d) to (%d, %d)\n",
		m.Scale,
		m.LeftX, m.LeftZ,
		m.RightX, m.RightZ,
	)

	err = mapping.Update(m)
	if err != nil {
		logrus.WithError(err).Fatal("Unable to save map")
	}

}
