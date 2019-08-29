// Package upstreams defines how to check version info in upstream repositories.
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

// UpstreamBase only contains a flavour. "Concrete" upstreams each implement their own fields.
type UpstreamBase struct {
	Flavour UpstreamFlavour `yaml:"flavour"`
}

// LatestVersion will always return an error.
// UpstreamBase is only used to determine which actual upstream needs to be called, so it cannot return a sensible value
func (u *UpstreamBase) LatestVersion() (string, error) {
	return "", errors.New("Cannot determine latest version for UpstreamBase")
}

// UpstreamFlavour is an enum of all supported upstreams and their string representation
type UpstreamFlavour string

const (
	// GithubFlavour is for Github releases
	GithubFlavour UpstreamFlavour = "github"
	// AMIFlavour is for Amazon Machine Images
	AMIFlavour UpstreamFlavour = "ami"
	// HelmFlavour is for Helm charts
	HelmFlavour UpstreamFlavour = "helm"
	// DummyFlavour is for testing
	DummyFlavour UpstreamFlavour = "dummy"
)
