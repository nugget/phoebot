package hooks

import (
	"fmt"
	"regexp"

	"github.com/nugget/phoebot/lib/discord"
	"github.com/nugget/phoebot/lib/products"

	"github.com/blang/semver"
	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

func RegVersion() (t Trigger) {
	t.Regexp = regexp.MustCompile("version report")
	t.Hook = ProcVersion
	t.Direct = true

	return t
}

func ProcVersion(dm *discordgo.MessageCreate) error {
	cutoff := semver.MustParse("0.0.0")
	logrus.WithField("cutoff", cutoff).Debug("Ignoring products older than this")

	mS := discordgo.MessageSend{}

	mE := discordgo.MessageEmbed{}
	mE.Description = "Current Minecraft Versions:"

	mE.Fields = make([]*discordgo.MessageEmbedField, 0)

	productList, err := products.GetImportant()
	if err != nil {
		logrus.WithError(err).Error("ProcVersion GetImportant Failed")
		return err
	}

	for _, p := range productList {
		mE.Fields = append(mE.Fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%s %s", p.Class, p.Name),
			Value:  fmt.Sprintf("%s", p.Latest.Version),
			Inline: true,
		})
	}

	mS.Embed = &mE

	if len(mE.Fields) > 0 {
		discord.Session.ChannelMessageSendComplex(dm.ChannelID, &mS)
	} else {
		discord.Session.ChannelMessageSend(dm.ChannelID, "I haven't seen any new versions lately, sorry. Try again later.")
	}
	return nil
}
