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

package dependency

import (
	"fmt"
	"strings"

	"github.com/blang/semver/v4"
	log "github.com/sirupsen/logrus"
)

// Version is the internal representation of a Version as a string and a scheme.
type Version struct {
	Version string
	Scheme  VersionScheme
}

// VersionScheme informs us on how to compare two versions.
type VersionScheme string

const (
	// Semver [Semantic versioning](https://semver.org/), default.
	Semver VersionScheme = "semver"
	// Alpha Alphanumeric, will use standard string sorting.
	Alpha VersionScheme = "alpha"
	// Random when releases do not support sorting (e.g. hashes).
	Random VersionScheme = "random"
)

type VersionUpdateInfo struct {
	Name            string
	Current         Version
	Latest          Version
	UpdateAvailable bool
}

// VersionUpdate represents the schema of the output format
// The output format is dictated by exportOptions.outputFormat.
type VersionUpdate struct {
	Name       string `json:"name"        yaml:"name"`
	Version    string `json:"version"     yaml:"version"`
	NewVersion string `json:"new_version" yaml:"new_version"`
}

// VersionSensitivity informs us on how to compare whether a version is more
// recent than another, for example to only notify on new major versions
// Only applicable to Semver versioning.
type VersionSensitivity string

const (
	// Patch version, e.g. 1.1.1 -> 1.1.2, default.
	Patch VersionSensitivity = "patch"
	// Minor version, e.g. 1.1.1 -> 1.2.0.
	Minor VersionSensitivity = "minor"
	// Major version, e.g. 1.1.1 -> 2.0.0.
	Major VersionSensitivity = "major"
)

// MoreRecentThan checks whether a given version is more recent than another one.
//
// If the VersionScheme is "random", then it will return true if a != b.
func (a Version) MoreRecentThan(b Version) (bool, error) {
	return a.MoreSensitivelyRecentThan(b, Patch)
}

// MoreSensitivelyRecentThan checks whether a given version is more recent than
// another one, accepting a VersionSensitivity argument
//
// If the VersionScheme is "random", then it will return true if a != b.
func (a Version) MoreSensitivelyRecentThan(b Version, sensitivity VersionSensitivity) (bool, error) {
	// Default to a Patch-level sensitivity
	if sensitivity == "" {
		sensitivity = Patch
	}

	if a.Scheme != b.Scheme {
		return false, fmt.Errorf("trying to compare incompatible 'Version' schemes: %s and %s", a.Scheme, b.Scheme)
	}

	switch a.Scheme {
	case Semver:
		aSemver, err := semver.ParseTolerant(a.Version)
		if err != nil {
			log.Debugf("Failed to semver-parse %s", a.Version)
			return false, err
		}

		bSemver, err := semver.ParseTolerant(b.Version)
		if err != nil {
			log.Debugf("Failed to semver-parse %s", b.Version)
			return false, err
		}

		return semverCompare(aSemver, bSemver, sensitivity)
	case Alpha:
		// Alphanumeric comparison (basic string compare)
		return a.Version > b.Version, nil
	case Random:
		// When identifiers are random (e.g. hashes), the newer version will just be a different version
		return a.Version != b.Version, nil
	default:
		return false, fmt.Errorf("unknown version scheme: %s", a.Scheme)
	}
}

// semverCompare compares two semver versions depending on a sensitivity level.
func semverCompare(a, b semver.Version, sensitivity VersionSensitivity) (bool, error) {
	switch sensitivity {
	case Major:
		return a.Major > b.Major, nil
	case Minor:
		return a.Major > b.Major || (a.Major == b.Major && a.Minor > b.Minor), nil
	case Patch:
		return a.GT(b), nil
	default:
		return false, fmt.Errorf("unknown version sensitivity: %s", sensitivity)
	}
}

// formatVersion preserves the string formatting from the template and ensures the version
// uses the same style (v-prefix).
func formatVersion(template, version string) string {
	// Use same prefix for both versions
	if strings.HasPrefix(template, "v") {
		if strings.HasPrefix(version, "v") {
			return version
		}
		return "v" + version
	}
	return strings.TrimPrefix(version, "v")
}
