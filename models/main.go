package models

import (
	"regexp"
	"time"

	"github.com/blang/semver"
	"github.com/bwmarrin/discordgo"
)

type LatestVersionFunction func(string) (semver.Version, error)
type RegisterFunction func() (string, LatestVersionFunction)
type GetTypesFunction func() ([]string, error)

type Subscription struct {
	ChannelID string `xml:"channelID"`
	Class     string `xml:"class"`
	Name      string `xml:"name"`
	Target    string `xml:"target"`
}

type SubChannel struct {
	Operation string
	Sub       Subscription
}

type LatestVersion struct {
	Version semver.Version `xml:"version"`
	Time    time.Time      `xml:"time"`
}

type DiscordMessage struct {
	ChannelID string
	Message   string
}

type Announcement struct {
	Product Product
	Message string
}

type Product struct {
	Class    string                `xml:"class"`
	Name     string                `xml:"type"`
	Latest   LatestVersion         `xml:"lastCheck"`
	Function LatestVersionFunction `xml:"-"`
}

type HookFunction func(*discordgo.MessageCreate) error

type Trigger struct {
	Regexp *regexp.Regexp
	Hook   HookFunction
	Direct bool
}
