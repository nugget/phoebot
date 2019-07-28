//
// Copyright (c) 2019 David McNett.  All Rights Reserved.
//
// SPDX-License-Identifier: BSD-2-Clause
//

package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nugget/phoebot/lib/builddata"
	"github.com/nugget/phoebot/lib/ipc"
	"github.com/nugget/phoebot/lib/mcserver"
	"github.com/nugget/phoebot/lib/phoelib"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func tsLine(buf string) {
	timeStamp := time.Now().Format("Mon 2-Jan 15:04:05")
	fmt.Printf("%s: %s\n", timeStamp, buf)
}

func shutdown() {
	logrus.Infof("Shutting down...")
	os.Exit(0)
}

func setupConfig() *viper.Viper {
	c := viper.New()
	c.AutomaticEnv()
	return c
}

func processChatStream(s mcserver.Server) {
	for {
		c, stillOpen := <-ipc.ServerChatStream

		logrus.WithFields(logrus.Fields{
			"message":   c,
			"stillOpen": stillOpen,
		}).Trace("serverChatStream message received")

		noColorsMessage := c.ClearString()

		for i, e := range c.Extra {
			logrus.WithFields(logrus.Fields{
				"i":             i,
				"text":          e.Text,
				"bold":          e.Bold,
				"italic":        e.Italic,
				"underlined":    e.UnderLined,
				"strikethrough": e.StrikeThrough,
				"obfuscated":    e.Obfuscated,
				"color":         e.Color,
			}).Trace("onChatMsg Debug Extra")
		}

		for i, w := range c.With {
			logrus.WithFields(logrus.Fields{
				"i": i,
				"w": string(w),
			}).Trace("onChatMsg Debug With")
		}

		f := s.LogFields(logrus.Fields{
			"event":     "processChatStream",
			"translate": c.Translate,
			"class":     mcserver.ChatMsgClass(c),
		})

		logrus.WithFields(f).Debug(noColorsMessage)
		tsLine(noColorsMessage)

		if !stillOpen {
			return
		}
	}
}

func processConsole(s mcserver.Server) {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		s.Client.Chat(scanner.Text())
	}
}

func main() {
	config := setupConfig()
	builddata.LogConversational()

	debugLevel := config.GetString("PHOEBOT_DEBUG")

	if debugLevel != "" {
		_, err := phoelib.LogLevel(debugLevel)
		if err != nil {
			logrus.WithError(err).Error("Unable to set LogLevel")
		}
	}

	ipc.InitStreams()

	mc, err := mcserver.New()
	if err != nil {
		logrus.WithError(err).Fatal("Fatal mcserver error")
	}

	err = mc.Authenticate(
		config.GetString("MINECRAFT_SERVER"),
		config.GetInt("MINECRAFT_PORT"),
		config.GetString("MOJANG_EMAIL"),
		config.GetString("MOJANG_PASSWORD"),
	)
	if err != nil {
		logrus.WithError(err).Error("Error with mcserver Authenticate")
	}

	go processChatStream(mc)
	go processConsole(mc)

	go mc.Handler()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	shutdown()
}
