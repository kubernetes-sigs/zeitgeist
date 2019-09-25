package dependencies

import (
	"fmt"

	"github.com/blang/semver"
)

// Version is the internal representation of a Version as a string and a scheme
type Version struct {
	Version string
	Scheme  VersionScheme
}

// VersionScheme informs us on how to compare two versions
type VersionScheme string

const (
	// Semver [Semantic versioning](https://semver.org/), default
	Semver VersionScheme = "semver"
	// Alpha Alphanumeric, will use standard string sorting
	Alpha VersionScheme = "alpha"
	// Random when releases do not support sorting (e.g. hashes)
	Random VersionScheme = "random"
)

// VersionSensitivity informs us on how to compare whether a version is more recent than another, for example to only notify on new major versions
// Only applicable to Semver versioning
type VersionSensitivity string

const (
	// Patch version, e.g. 1.1.1 -> 1.1.2, default
	Patch VersionSensitivity = "patch"
	// Minor version, e.g. 1.1.1 -> 1.2.0
	Minor VersionSensitivity = "minor"
	// Major version, e.g. 1.1.1 -> 2.0.0
	Major VersionSensitivity = "major"
)

// MoreRecentThan checks whether a given version is more recent than another one.
//
// If the VersionScheme is "random", then it will return true if a != b.
func (a Version) MoreRecentThan(b Version) (bool, error) {
	return a.MoreSensitivelyRecentThan(b, Patch)
}

// MoreSensitivelyRecentThan checks whether a given version is more recent than another one, accepting a VersionSensitivity argument
//
// If the VersionScheme is "random", then it will return true if a != b.
func (a Version) MoreSensitivelyRecentThan(b Version, sensitivity VersionSensitivity) (bool, error) {
	// Default to a Patch-level sensitivity
	if sensitivity == "" {
		sensitivity = Patch
	}
	if a.Scheme != b.Scheme {
		return false, fmt.Errorf("Trying to compare incompatible Version schemes: %s and %s", a.Scheme, b.Scheme)
	}
	switch a.Scheme {
	case Semver:
		aSemver, err := semver.Parse(string(a.Version))
		if err != nil {
			return false, err
		}
		bSemver, err := semver.Parse(string(b.Version))
		if err != nil {
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
		return false, fmt.Errorf("Unknown version scheme: %s", a.Scheme)
	}
}

// semverCompare compares two semver versions depending on a sensitivity level
func semverCompare(a semver.Version, b semver.Version, sensitivity VersionSensitivity) (bool, error) {
	switch sensitivity {
	case Major:
		return a.Major > b.Major, nil
	case Minor:
		return !(a.Major < b.Major) && a.Minor > b.Minor, nil
	case Patch:
		return a.GT(b), nil
	default:
		return false, fmt.Errorf("Unknown version sensitivity: %s", sensitivity)
	}
}
