package main

import (
	"fmt"

	"github.com/nugget/phoebot/serverpro"

	"github.com/bwmarrin/discordgo"
)

func main() {
	dg, err := discordgo.New("Bot " + os.Env("DISCORD_BOT_TOKEN"))
	if err != nil {
		logger.Fatal(err)
	}

	lv, err := serverpro.LatestVersion("Paper")

	fmt.Println(lv, err)
}
