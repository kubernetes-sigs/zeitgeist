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
	"testing"

	"github.com/stretchr/testify/require"
)

// TODO: These tests should be refactored to be table-driven
//       Additionally, we can use https://github.com/stretchr/testify/require
//       to check the various statuses.

func TestSanity(t *testing.T) {
	var err error

	a := Version{"1.0.0", Semver}
	b := Version{"2.0.0", Alpha}

	_, err = a.MoreRecentThan(b)
	require.NotNil(t, err)

	a = Version{"1.0.0", "Foo"}
	b = Version{"2.0.0", "Foo"}

	_, err = a.MoreRecentThan(b)
	require.NotNil(t, err)

	a = Version{"ami-1234", Semver}
	b = Version{"ami-4567", Semver}

	_, err = a.MoreRecentThan(b)
	require.NotNil(t, err)

	a = Version{"1.0.0", Semver}
	b = Version{"bad-version", Semver}

	_, err = a.MoreRecentThan(b)
	require.NotNil(t, err)
}

func TestSemverVersions(t *testing.T) {
	a := Version{"1.0.0", Semver}
	b := Version{"2.0.0", Semver}

	// nolint: errcheck
	shouldBeFalse, _ := a.MoreRecentThan(b)
	require.False(t, shouldBeFalse)

	// nolint: errcheck
	shouldBeTrue, _ := b.MoreRecentThan(a)
	require.True(t, shouldBeTrue)

	// nolint: errcheck
	shouldBeFalse, _ = a.MoreRecentThan(a)
	require.False(t, shouldBeFalse)
}

func TestSemverSensitiveVersions(t *testing.T) {
	a := Version{"1.0.0", Semver}
	b := Version{"1.1.0", Semver}

	// nolint: errcheck
	shouldBeFalse, _ := b.MoreSensitivelyRecentThan(a, Major)
	require.False(t, shouldBeFalse)

	// nolint: errcheck
	shouldBeTrue, _ := b.MoreSensitivelyRecentThan(a, Minor)
	require.True(t, shouldBeTrue)

	// nolint: errcheck
	shouldBeTrue, _ = b.MoreSensitivelyRecentThan(a, Patch)
	require.True(t, shouldBeTrue)

	a = Version{"1.0.0", Semver}
	b = Version{"1.0.1", Semver}

	// nolint: errcheck
	shouldBeFalse, _ = b.MoreSensitivelyRecentThan(a, Major)
	require.False(t, shouldBeFalse)

	// nolint: errcheck
	shouldBeFalse, _ = b.MoreSensitivelyRecentThan(a, Minor)
	require.False(t, shouldBeFalse)

	// nolint: errcheck
	shouldBeTrue, _ = b.MoreSensitivelyRecentThan(a, Patch)
	require.True(t, shouldBeTrue)

	_, shouldError := b.MoreSensitivelyRecentThan(a, "foo")
	require.Error(t, shouldError)

	a = Version{"6.21.0", Semver}
	b = Version{"8.1.8", Semver}

	// nolint: errcheck
	shouldBeTrue, _ = b.MoreSensitivelyRecentThan(a, Minor)
	require.True(t, shouldBeTrue)
}

func TestAlphaVersions(t *testing.T) {
	a := Version{"20180101-commitid", Alpha}
	b := Version{"20180505-commitid", Alpha}

	// nolint: errcheck
	shouldBeFalse, _ := a.MoreRecentThan(b)
	require.False(t, shouldBeFalse)

	// nolint: errcheck
	shouldBeTrue, _ := b.MoreRecentThan(a)
	require.True(t, shouldBeTrue)

	// nolint: errcheck
	shouldBeFalse, _ = a.MoreRecentThan(a)
	require.False(t, shouldBeFalse)
}

func TestRandomVersions(t *testing.T) {
	a := Version{"ami-09bbefc07310f7914", Random}
	b := Version{"ami-0199284372364b02a", Random}

	// nolint: errcheck
	shouldBeTrue, _ := b.MoreRecentThan(a)
	require.True(t, shouldBeTrue)

	// nolint: errcheck
	shouldBeFalse, _ := a.MoreRecentThan(a)
	require.False(t, shouldBeFalse)
}
