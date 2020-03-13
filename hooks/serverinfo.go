package hooks

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/nugget/phoebot/lib/console"
	"github.com/nugget/phoebot/lib/discord"
	"github.com/sirupsen/logrus"

	"github.com/bwmarrin/discordgo"
)

func RegServerInfo() (t Trigger) {
	t.Regexp = regexp.MustCompile("(?i)!server ?info")
	t.Hook = ProcServerInfo
	t.Direct = false

	return t
}

func ProcServerInfo(dm *discordgo.MessageCreate) error {
	si, err := console.GetServerInfo()
	if err != nil {
		logrus.WithError(err).Error("ProcServerInfo failure")
		return err
	}

	mS := discordgo.MessageSend{}
	mE := discordgo.MessageEmbed{}

	mF := discordgo.MessageEmbedFooter{}
	mF.Text = fmt.Sprintf("%s", console.Hostname())
	mE.Footer = &mF

	mE.Fields = make([]*discordgo.MessageEmbedField, 0)

	mE.Fields = append(mE.Fields, &discordgo.MessageEmbedField{
		Name: fmt.Sprintf("Players (%d/%d)",
			len(si.Players),
			si.MaxPlayers,
		),
		Value:  strings.Join(si.Players, ", "),
		Inline: false,
	})

	mE.Fields = append(mE.Fields, &discordgo.MessageEmbedField{
		Name:   "Version",
		Value:  si.Version,
		Inline: false,
	})

	mE.Fields = append(mE.Fields, &discordgo.MessageEmbedField{
		Name:   "Performance (1m, 5m, 15m)",
		Value:  fmt.Sprintf("%2.2f, %2.2f, %2.2f TPS", si.Tps1, si.Tps5, si.Tps15),
		Inline: false,
	})

	mS.Embed = &mE

	discord.Session.ChannelMessageSendComplex(dm.ChannelID, &mS)

	return nil
}

func RegServerList() (t Trigger) {
	t.Regexp = regexp.MustCompile("(?i)^(who is on the server|!who|!list)")
	t.Hook = ProcServerList
	t.Direct = false

	return t
}

func ProcServerList(dm *discordgo.MessageCreate) error {
	si, err := console.GetServerInfo()
	if err != nil {
		logrus.WithError(err).Error("ProcServerList failure")
		return err
	}

	mS := discordgo.MessageSend{}
	mE := discordgo.MessageEmbed{}

	mF := discordgo.MessageEmbedFooter{}
	mF.Text = fmt.Sprintf("%s", console.Hostname())
	mE.Footer = &mF

	mE.Fields = make([]*discordgo.MessageEmbedField, 0)

	mE.Fields = append(mE.Fields, &discordgo.MessageEmbedField{
		Name: fmt.Sprintf("Players (%d/%d)",
			len(si.Players),
			si.MaxPlayers,
		),
		Value:  strings.Join(si.Players, ", "),
		Inline: false,
	})

	mS.Embed = &mE

	discord.Session.ChannelMessageSendComplex(dm.ChannelID, &mS)

	return nil
}
