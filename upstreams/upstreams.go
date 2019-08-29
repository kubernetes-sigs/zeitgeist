package upstreams

import (
	"errors"
)

// The base Upstream struct only contains a flavour. "Concrete" upstreams each implement their own fields.
type UpstreamBase struct {
	Flavour UpstreamFlavour `yaml:"flavour"`
}

// This function will always return an error.
// UpstreamBase is only used to determine which actual upstream needs to be called, so it cannot return a sensible value
func (u *UpstreamBase) LatestVersion() (string, error) {
	return "", errors.New("Cannot determine latest version for UpstreamBase")
}

// All supported upstreams and their string representation
type UpstreamFlavour string

const (
	GithubFlavour UpstreamFlavour = "github"
	AMIFlavour    UpstreamFlavour = "ami"
	DummyFlavour  UpstreamFlavour = "dummy"
)
