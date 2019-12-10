package models

import (
	"time"

	"github.com/blang/semver"
)

type LatestVersionFunction func(string) (semver.Version, error)
type RegisterFunction func() (string, LatestVersionFunction)
type GetTypesFunction func() ([]string, error)

type Subscription struct {
	ChannelID string
	Class     string
	Name      string
	Target    string
}

type LatestVersion struct {
	Version semver.Version
	Time    time.Time
}

type DiscordMessage struct {
	ChannelID string
	Message   string
}

type Whisper struct {
	Who     string
	Message string
}

type Product struct {
	Class    string
	Name     string
	Latest   LatestVersion
	Function LatestVersionFunction
}

type Article struct {
	Title       string
	URL         string
	PublishDate time.Time
	Release     bool
	Product     string
	Version     string
}
