package upstreams

import (
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"
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

// Custom unmarshalling of Upstream to add extra validation
func (decoded *UpstreamBase) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Use a different type to prevent infinite loop in unmarshalling
	type UpstreamYAML UpstreamBase
	u := (*UpstreamYAML)(decoded)
	if err := unmarshal(&u); err != nil {
		return err
	}

	if u.Flavour == "" {
		return fmt.Errorf("Upstream has no `flavour`: %v", u)
	}
	// Validate flavours and return
	switch u.Flavour {
	case DummyFlavour, GithubFlavour, AMIFlavour:
		log.Debugf("Deserialised Upstream: %v", u)
		return nil
	default:
		return fmt.Errorf("Unknown upstream flavour: %s", u.Flavour)
	}
}
