package main

import (
	"github.com/nugget/phoebot/functions"

	"github.com/bwmarrin/discordgo"
)

type hookFunction func(*discordgo.MessageCreate) error

func LoadTriggers() error {
	// This is the baseline feature you can use to pattern new features you
	// want to add
	triggers = append(triggers, functions.RegTemplate())
	triggers = append(triggers, functions.RegLoglevel())

	triggers = append(triggers, functions.RegSubscriptions())
	triggers = append(triggers, functions.RegListSubscriptions())
	triggers = append(triggers, functions.RegVersion())
	triggers = append(triggers, functions.RegTimezones())
	triggers = append(triggers, functions.RegStatus())

	return nil
}
