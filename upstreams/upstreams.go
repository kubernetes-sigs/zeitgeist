package upstreams

import (
	"errors"
)

type UpstreamBase struct {
	Flavour     UpstreamFlavour `yaml:"flavour"`
	URL         string          `yaml:"url"`
	Constraints string          `yaml:"constraints"`
}

func (u *UpstreamBase) LatestVersion() (string, error) {
	return "", errors.New("Cannot determine latest version for UpstreamBase")
}

type UpstreamFlavour string

const (
	GithubFlavour UpstreamFlavour = "github"
	AMIFlavour    UpstreamFlavour = "ami"
	DummyFlavour  UpstreamFlavour = "dummy"
)
