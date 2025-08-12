/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package upstream defines how to check version info in upstream repositories.
//
// Upstream types are identified by their _flavour_, represented as a string (see Flavour).
//
// Different Upstream types can have their own parameters, but they must:
//
//   - Include the BaseUpstream type
//   - Define a LatestVersion() function that returns the latest available version as a string
package upstream

import (
	"errors"

	"github.com/blang/semver/v4"
	log "github.com/sirupsen/logrus"

	"sigs.k8s.io/release-utils/helpers"
)

// Base only contains a flavour. "Concrete" upstreams each implement their own fields.
type Base struct {
	Flavour Flavour `yaml:"flavour"`
}

// LatestVersion will always return an error.
// Base is only used to determine which actual upstream needs to be called, so it cannot return a sensible value.
func (u *Base) LatestVersion() (string, error) {
	return "", errors.New("cannot determine latest version for Base")
}

// Flavour is an enum of all supported upstreams and their string representation.
type Flavour string

const (
	// GithubFlavour is for Github releases.
	GithubFlavour Flavour = "github"

	// GitLabFlavour is for GitLab releases.
	GitLabFlavour Flavour = "gitlab"

	// AMIFlavour is for Amazon Machine Images.
	AMIFlavour Flavour = "ami"

	// HelmFlavour is for Helm Charts.
	HelmFlavour Flavour = "helm"

	// ContainerFlavour is for Container Images.
	ContainerFlavour Flavour = "container"

	// EKSFlavour is for Elastic Kubernetes Service.
	EKSFlavour Flavour = "eks"

	// DummyFlavour is for testing.
	DummyFlavour Flavour = "dummy"

	DefaultSemVerConstraints = ">= 0.0.0"
)

func selectHighestVersion(constraints string, expectedRange semver.Range, tags []string) (string, error) {
	var candidateVersion semver.Version
	candidateVersionString := "" // keep the version string separately as it may contain a leading `v`
	for _, tag := range tags {
		// Try to match semver and range
		version, err := helpers.TagStringToSemver(tag)
		if err != nil {
			log.Debugf("Error parsing version %s (%#v) as semver, cannot validate semver constraints", tag, err)
			continue
		}

		if !expectedRange(version) {
			log.Debugf("Skipping release not matching range constraints (%s): %s", constraints, tag)
			continue
		}

		log.Debugf("Found potential release: %s\n", version.String())
		if candidateVersionString == "" || version.GT(candidateVersion) {
			log.Debugf("Release is the newest found so far: %s", version.String())
			candidateVersion = version
			candidateVersionString = tag
		}
	}

	if candidateVersionString != "" {
		return candidateVersionString, nil
	}

	// No latest version found – no versions? Only prereleases?
	return "", errors.New("no potential version found")
}
