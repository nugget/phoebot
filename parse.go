package main

import (
	"regexp"

	"github.com/bwmarrin/discordgo"
)

type hookFunction func(*discordgo.MessageCreate) error

type Trigger struct {
	Regexp *regexp.Regexp
	Hook   hookFunction
	Direct bool
}

func LoadTriggers() error {
	// This is the baseline feature you can use to pattern new features you
	// want to add
	triggers = append(triggers, regTemplate())
	triggers = append(triggers, regLoglevel())

	triggers = append(triggers, regSubscriptions())
	triggers = append(triggers, regVersion())
	triggers = append(triggers, regTimezones())
	triggers = append(triggers, regStatus())

	return nil
}
