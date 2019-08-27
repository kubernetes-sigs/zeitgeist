package upstreams

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestDummy(t *testing.T) {
	var u Dummy
	input := []byte("flavour: dummy")
	err := yaml.Unmarshal(input, &u)
	if err != nil {
		t.Errorf("Failed to deserialise valid yaml:\n%s\nError: %v", input, err)
	}
	v, err := u.LatestVersion()
	if err != nil {
		t.Errorf("Failed to get dummy latest version: %v", err)
	}
	if v != "1.0.0" {
		t.Errorf("Dummy latest version isn't 1.0.0: %s", v)
	}
}
