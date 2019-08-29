package dependencies

import (
	"fmt"

	"github.com/blang/semver"
)

// Internal representation of a Version as a string and a scheme
type Version struct {
	Version string
	Scheme  VersionScheme
}

// The VersionScheme informs us on how to compare whether a version is more recent than another
type VersionScheme string

const (
	// [Semantic versioning](https://semver.org/), default
	Semver VersionScheme = "semver"
	// Alphanumeric, will use standard string sorting
	Alpha VersionScheme = "alpha"
	// "Random": when releases do not support sorting (e.g. hashes)
	Random VersionScheme = "random"
)

// Checks whether a given version is more recent than another one.
//
// If the VersionScheme is "random", then it will return true if a != b.
func (a Version) MoreRecentThan(b Version) (bool, error) {
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
		return aSemver.GT(bSemver), nil
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
