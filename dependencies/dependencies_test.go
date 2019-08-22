package dependencies

import (
	"testing"
)

// Happy Path test
func TestLocal(t *testing.T) {
	err := LocalCheck("../testdata/local.yaml")
	if err != nil {
		t.Errorf("Happy path local test returned: %v", err)
	}
}

func TestRemote(t *testing.T) {
	err := RemoteCheck("../testdata/remote.yaml", "")
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
	err := RemoteCheck("../testdata/unknown-upstream.yaml", "")
	if err == nil {
		t.Errorf("Did not return an error when it should have")
	}
}
