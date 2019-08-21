package dependencies

import (
	"strings"

	"github.com/blang/semver"
)

type Version string

func (a Version) MoreRecentThan(b Version) bool {
	// Try and parse as Semver first
	semverComparison := true
	aSemver, err := semver.Parse(string(a))
	if err != nil {
		semverComparison = false
	}
	bSemver, err := semver.Parse(string(b))
	if err != nil {
		semverComparison = false
	}
	if semverComparison {
		return aSemver.GT(bSemver)
	} else {
		// Failed semver: fallback to standard string comparison (lexicographic)
		return strings.Compare(string(a), string(b)) > 0
	}
}
