package upstreams

import (
	"testing"
)

func TestAMIHappyPath(t *testing.T) {
	ami := AMI{
		Owner: "amazon",
		Name:  "amazon-eks-node-1.13-*",
	}
	latestVersion, err := ami.LatestVersion()
	if err != nil {
		t.Errorf("Failed AMI happy path test: %v", err)
	}
	if latestVersion == "" {
		t.Errorf("Got an empty latestVersion")
	}
}

func TestAMIDoesntExist(t *testing.T) {
	fakeAmi := "this-ami-doesnt-exist-zeitgeist"
	ami := AMI{
		Owner: "amazon",
		Name:  fakeAmi,
	}
	_, err := ami.LatestVersion()
	if err == nil {
		t.Errorf("Found a latest version for unknown AMI: %s", fakeAmi)
	}
}
