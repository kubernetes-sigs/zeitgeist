package dependencies

import (
	"testing"
)

func TestSemverVersions(t *testing.T) {
	a := Version("1.0.0")
	b := Version("2.0.0")
	shouldBeFalse := a.MoreRecentThan(b)
	if shouldBeFalse == true {
		t.Errorf("Version %v is more recent that version %v; should be the opposite", a, b)
	}
	shouldBeTrue := b.MoreRecentThan(a)
	if shouldBeTrue == false {
		t.Errorf("Version %v is more recent that version %v; should be the opposite", b, a)
	}
	shouldBeFalse = a.MoreRecentThan(a)
	if shouldBeFalse == true {
		t.Errorf("Version %v is more recent than itself; it should not be", a)
	}
}

func TestNonSemverVersions(t *testing.T) {
	a := Version("20180101-commitid")
	b := Version("20180501-commitid")
	shouldBeFalse := a.MoreRecentThan(b)
	if shouldBeFalse == true {
		t.Errorf("Version %v is more recent that version %v; should be the opposite", a, b)
	}
	shouldBeTrue := b.MoreRecentThan(a)
	if shouldBeTrue == false {
		t.Errorf("Version %v is more recent that version %v; should be the opposite", b, a)
	}
	shouldBeFalse = a.MoreRecentThan(a)
	if shouldBeFalse == true {
		t.Errorf("Version %v is more recent than itself; it should not be", a)
	}
}
