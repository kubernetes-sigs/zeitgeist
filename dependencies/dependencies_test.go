package dependencies

import (
	"testing"

	"gopkg.in/yaml.v3"
)

// Happy Path test
func TestLocal(t *testing.T) {
	err := LocalCheck("../testdata/local.yaml")
	if err != nil {
		t.Errorf("Happy path local test returned: %v", err)
	}
}

func TestRemote(t *testing.T) {
	err := RemoteCheck("../testdata/remote.yaml")
	if err != nil {
		t.Errorf("Happy path local test returned: %v", err)
	}
}

func TestDummyRemote(t *testing.T) {
	err := RemoteCheck("../testdata/remote-dummy.yaml")
	if err != nil {
		t.Errorf("Happy path local test returned: %v", err)
	}
}

func TestRemoteConstraint(t *testing.T) {
	err := RemoteCheck("../testdata/remote-constraint.yaml")
	if err != nil {
		t.Errorf("Happy path local test returned: %v", err)
	}
}

func TestBrokenFile(t *testing.T) {
	err := LocalCheck("../testdata/does-not-exist")
	if err == nil {
		t.Errorf("Did not return an error on trying to open a file that doesn't exist")
	}
	err = LocalCheck("../testdata/Dockerfile")
	if err == nil {
		t.Errorf("Did not return an error on trying to open a non-yaml file")
	}
}

func TestLocalOutOfSync(t *testing.T) {
	err := LocalCheck("../testdata/local-out-of-sync.yaml")
	if err == nil {
		t.Errorf("Did not return an error when it should have")
	}
}

func TestFileDoesntExist(t *testing.T) {
	err := LocalCheck("../testdata/local-no-file.yaml")
	if err == nil {
		t.Errorf("Did not return an error when it should have")
	}
}

func TestUnknownUpstreamFlavour(t *testing.T) {
	err := RemoteCheck("../testdata/unknown-upstream.yaml")
	if err == nil {
		t.Errorf("Did not return an error when it should have")
	}
}

func TestDeserialising(t *testing.T) {
	invalidYamls := []string{
		"a b c",
		"name:",
		"name: test",
		"version: 1.0.0",
	}

	for _, invalid := range invalidYamls {
		var d Dependency

		err := yaml.Unmarshal([]byte(invalid), &d)
		if err == nil {
			t.Errorf("Did not return an error when it should have on invalid yaml:\n\n%s\n", invalid)
		}
	}

	validYamls := []string{
		"name: test\nversion: 1.0.0",
		"name: test\nversion: 100",
	}

	for _, valid := range validYamls {
		var d Dependency
		err := yaml.Unmarshal([]byte(valid), &d)
		if err != nil {
			t.Errorf("Failed to deserialise valid yaml:\n%s", valid)
		}
	}
}
