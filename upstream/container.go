/*
Copyright 2021 The Kubernetes Authors.

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

package upstream

import (
	"sort"
	"strings"

	"github.com/blang/semver"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"sigs.k8s.io/zeitgeist/pkg/container"
)

// Container upstream representation
type Container struct {
	Base `mapstructure:",squash"`
	// Registry URL, e.g. gcr.io/k8s-staging-kubernetes/conformance
	Registry string
	// Optional: semver constraints, e.g. < 2.0.0
	// Will have no effect if the dependency does not follow Semver
	Constraints string
}

// LatestVersion returns the latest tag for the given repository
// (depending on the Constraints if set).
func (upstream Container) LatestVersion() (string, error) { // nolint:gocritic
	log.Debugf("Using Container flavour")
	return highestSemanticImageTag(&upstream)
}

func highestSemanticImageTag(upstream *Container) (string, error) {
	client := container.New()

	semverConstraints := upstream.Constraints
	if semverConstraints == "" {
		// If no range is passed, just use the broadest possible range
		semverConstraints = DefaultSemVerConstraints
	}
	expectedRange, err := semver.ParseRange(semverConstraints)
	if err != nil {
		return "", errors.Errorf("invalid semver constraints range: %v", upstream.Constraints)
	}

	log.Debugf("Retrieving tags for %s...", upstream.Registry)
	tags, err := client.ListTags(upstream.Registry)
	if err != nil {
		return "", errors.Wrap(err, "retrieving Container tags")
	}

	sort.Sort(sort.Reverse(sort.StringSlice(tags)))

	for _, tag := range tags {
		// Try to match semver and range
		version, err := semver.Parse(strings.Trim(tag, "v"))
		// Try to match semver and range
		if err != nil {
			log.Debugf("Error parsing version %v (%v) as semver, cannot validate semver constraints", tag, err)
		} else if !expectedRange(version) {
			log.Debugf("Skipping release not matching range constraints (%v): %v\n", upstream.Constraints, tag)
			continue
		}

		log.Debugf("Found latest matching tag: %v\n", version)
		return version.String(), nil
	}

	return "", errors.Errorf("no potential tag found")
}
