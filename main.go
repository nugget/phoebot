package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/nugget/phoebot/serverpro"
)

func main() {
	dg, err := discordgo.New("Bot " + os.Env("DISCORD_BOT_TOKEN"))
	if err != nil {
		logger.Fatal(err)
	}

	lv, err := serverpro.LatestVersion("Paper")

	fmt.Println(lv, err)
}
