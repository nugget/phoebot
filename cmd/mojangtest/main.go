package main

import (
	"github.com/nugget/phoebot/lib/db"
	"github.com/nugget/phoebot/lib/phoelib"
	"github.com/nugget/phoebot/products/mojang"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	config := viper.New()
	config.AutomaticEnv()

	phoelib.LogLevel("debug")
	dbURI := config.GetString("DATABASE_URI")

	err := db.Connect(dbURI)
	if err != nil {
		logrus.WithError(err).Fatal("Unable to connect to database")
	}

	logrus.Info("Mojang Test Harness")

	err = mojang.SeekReleases()
	if err != nil {
		logrus.WithError(err).Error("SeekReleases failed")
	}
}
