package dependencies

import (
	"testing"
)

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
	shouldBeFalse, _ := a.MoreRecentThan(b)
	if shouldBeFalse == true {
		t.Errorf("Version %v is more recent that version %v; should be the opposite", a, b)
	}
	shouldBeTrue, _ := b.MoreRecentThan(a)
	if shouldBeTrue == false {
		t.Errorf("Version %v is more recent that version %v; should be the opposite", b, a)
	}
	shouldBeFalse, _ = a.MoreRecentThan(a)
	if shouldBeFalse == true {
		t.Errorf("Version %v is more recent than itself; it should not be", a)
	}
}

func TestAlphaVersions(t *testing.T) {
	a := Version{"20180101-commitid", Alpha}
	b := Version{"20180505-commitid", Alpha}
	shouldBeFalse, _ := a.MoreRecentThan(b)
	if shouldBeFalse == true {
		t.Errorf("Version %v is more recent that version %v; should be the opposite", a, b)
	}
	shouldBeTrue, _ := b.MoreRecentThan(a)
	if shouldBeTrue == false {
		t.Errorf("Version %v is more recent that version %v; should be the opposite", b, a)
	}
	shouldBeFalse, _ = a.MoreRecentThan(a)
	if shouldBeFalse == true {
		t.Errorf("Version %v is more recent than itself; it should not be", a)
	}
}

func TestRandomVersions(t *testing.T) {
	a := Version{"ami-09bbefc07310f7914", Random}
	b := Version{"ami-0199284372364b02a", Random}
	shouldBeTrue, _ := b.MoreRecentThan(a)
	if shouldBeTrue == false {
		t.Errorf("Version %v is more recent that version %v; should be the opposite", b, a)
	}
	shouldBeFalse, _ := a.MoreRecentThan(a)
	if shouldBeFalse == true {
		t.Errorf("Version %v is more recent than itself; it should not be", a)
	}
}
