package models

import (
	"github.com/blang/semver"
)

type LatestVersionFunction func(string) (semver.Version, error)
type RegisterFunction func() (string, LatestVersionFunction)
type GetTypesFunction func() ([]string, error)
