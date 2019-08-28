package dependencies

import (
	"fmt"

	"github.com/blang/semver"
)

type Version struct {
	Version string
	Scheme  VersionScheme
}

type VersionScheme string

const (
	Semver VersionScheme = "semver"
	Alpha  VersionScheme = "alpha"
	Random VersionScheme = "random"
)

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
