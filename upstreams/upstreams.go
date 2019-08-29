// Defines how to check version info in upstream repositories.
//
// Upstream types are identified by their _flavour_, represented as a string (see UpstreamFlavour).
//
// Different Upstream types can have their own parameters, but they must:
//
//	- Include the BaseUpstream type
//	- Define a LatestVersion() function that returns the latest available version as a string
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
	HelmFlavour   UpstreamFlavour = "helm"
	DummyFlavour  UpstreamFlavour = "dummy"
)
