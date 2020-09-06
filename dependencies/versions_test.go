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

package dependencies

import (
	"testing"
)

// TODO: These tests should be refactored to be table-driven
//       Additionally, we can use https://github.com/stretchr/testify/require
//       to check the various statuses.

func TestSanity(t *testing.T) {
	var err error

	a := Version{"1.0.0", Semver}
	b := Version{"2.0.0", Alpha}

	_, err = a.MoreRecentThan(b)
	if err == nil {
		t.Errorf("Should error on copmparing different types")
	}

	a = Version{"1.0.0", "Foo"}
	b = Version{"2.0.0", "Foo"}

	_, err = a.MoreRecentThan(b)
	if err == nil {
		t.Errorf("Should error on copmparing unknown types")
	}

	a = Version{"ami-1234", Semver}
	b = Version{"ami-4567", Semver}

	_, err = a.MoreRecentThan(b)
	if err == nil {
		t.Errorf("Should error on broken Semver strings")
	}

	a = Version{"1.0.0", Semver}
	b = Version{"bad-version", Semver}

	_, err = a.MoreRecentThan(b)
	if err == nil {
		t.Errorf("Should error on broken Semver strings")
	}
}

func TestSemverVersions(t *testing.T) {
	a := Version{"1.0.0", Semver}
	b := Version{"2.0.0", Semver}

	// nolint: errcheck
	shouldBeFalse, _ := a.MoreRecentThan(b)
	if shouldBeFalse == true {
		t.Errorf("Version %v is more recent that version %v; should be the opposite", a, b)
	}

	// nolint: errcheck
	shouldBeTrue, _ := b.MoreRecentThan(a)
	if shouldBeTrue == false {
		t.Errorf("Version %v is more recent that version %v; should be the opposite", b, a)
	}

	// nolint: errcheck
	shouldBeFalse, _ = a.MoreRecentThan(a)
	if shouldBeFalse == true {
		t.Errorf("Version %v is more recent than itself; it should not be", a)
	}
}

func TestSemverSensitiveVersions(t *testing.T) {
	a := Version{"1.0.0", Semver}
	b := Version{"1.1.0", Semver}

	// nolint: errcheck
	shouldBeFalse, _ := b.MoreSensitivelyRecentThan(a, Major)
	if shouldBeFalse == true {
		t.Errorf("Version %v should not be more recent that version %v with sensitivity %v", b, a, Major)
	}

	// nolint: errcheck
	shouldBeTrue, _ := b.MoreSensitivelyRecentThan(a, Minor)
	if shouldBeTrue != true {
		t.Errorf("Version %v should be more recent that version %v with sensitivity %v", b, a, Minor)
	}

	// nolint: errcheck
	shouldBeTrue, _ = b.MoreSensitivelyRecentThan(a, Patch)
	if shouldBeTrue != true {
		t.Errorf("Version %v should be more recent that version %v with sensitivity %v", b, a, Patch)
	}

	a = Version{"1.0.0", Semver}
	b = Version{"1.0.1", Semver}

	// nolint: errcheck
	shouldBeFalse, _ = b.MoreSensitivelyRecentThan(a, Major)
	if shouldBeFalse == true {
		t.Errorf("Version %v should not be more recent that version %v with sensitivity %v", b, a, Major)
	}

	// nolint: errcheck
	shouldBeFalse, _ = b.MoreSensitivelyRecentThan(a, Minor)
	if shouldBeFalse == true {
		t.Errorf("Version %v should not be more recent that version %v with sensitivity %v", b, a, Minor)
	}

	// nolint: errcheck
	shouldBeTrue, _ = b.MoreSensitivelyRecentThan(a, Patch)
	if shouldBeTrue != true {
		t.Errorf("Version %v should be more recent that version %v with sensitivity %v", b, a, Patch)
	}

	_, shouldError := b.MoreSensitivelyRecentThan(a, "foo")
	if shouldError == nil {
		t.Errorf("Should error on sensitivity %v", "foo")
	}

	a = Version{"6.21.0", Semver}
	b = Version{"8.1.8", Semver}

	// nolint: errcheck
	shouldBeTrue, _ = b.MoreSensitivelyRecentThan(a, Minor)
	if shouldBeTrue != true {
		t.Errorf("Version %v should be more recent that version %v with sensitivity %v", b, a, Minor)
	}
}

func TestAlphaVersions(t *testing.T) {
	a := Version{"20180101-commitid", Alpha}
	b := Version{"20180505-commitid", Alpha}

	// nolint: errcheck
	shouldBeFalse, _ := a.MoreRecentThan(b)
	if shouldBeFalse == true {
		t.Errorf("Version %v is more recent that version %v; should be the opposite", a, b)
	}

	// nolint: errcheck
	shouldBeTrue, _ := b.MoreRecentThan(a)
	if shouldBeTrue == false {
		t.Errorf("Version %v is more recent that version %v; should be the opposite", b, a)
	}

	// nolint: errcheck
	shouldBeFalse, _ = a.MoreRecentThan(a)
	if shouldBeFalse == true {
		t.Errorf("Version %v is more recent than itself; it should not be", a)
	}
}

func TestRandomVersions(t *testing.T) {
	a := Version{"ami-09bbefc07310f7914", Random}
	b := Version{"ami-0199284372364b02a", Random}

	// nolint: errcheck
	shouldBeTrue, _ := b.MoreRecentThan(a)
	if shouldBeTrue == false {
		t.Errorf("Version %v is more recent that version %v; should be the opposite", b, a)
	}

	// nolint: errcheck
	shouldBeFalse, _ := a.MoreRecentThan(a)
	if shouldBeFalse == true {
		t.Errorf("Version %v is more recent than itself; it should not be", a)
	}
}
