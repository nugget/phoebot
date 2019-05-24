package models

import (
	"fmt"
	"time"

	"github.com/blang/semver"
)

type LatestVersionFunction func(string) (semver.Version, error)
type RegisterFunction func() (string, LatestVersionFunction)
type GetTypesFunction func() ([]string, error)

type Subscription struct {
	ChannelID string  `xml:"channelID"`
	Product   Product `xml:"product"`
	Target    string  `xml:"target"`
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

func (p Product) String() string {
	return fmt.Sprintf("%s:%s", p.Class, p.Name)
}
