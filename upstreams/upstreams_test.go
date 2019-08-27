package upstreams

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestDeserialising(t *testing.T) {
	invalidYamls := []string{
		"a b c",
		"flavour:",
		"flavour: unknown-test",
	}

	for _, invalid := range invalidYamls {
		var u UpstreamBase

		err := yaml.Unmarshal([]byte(invalid), &u)
		if err == nil {
			t.Errorf("Did not return an error when it should have on invalid yaml:\n\n%s\n", invalid)
		}
	}

	validYamls := []string{
		"flavour: dummy",
		"flavour: github",
	}

	for _, valid := range validYamls {
		var u UpstreamBase
		err := yaml.Unmarshal([]byte(valid), &u)
		if err != nil {
			t.Errorf("Failed to deserialise valid yaml:\n%s", valid)
		}
	}
}

func TestUpstreamBaseLatestVersion(t *testing.T) {
	var u UpstreamBase
	input := []byte("flavour: dummy")
	err := yaml.Unmarshal(input, &u)
	if err != nil {
		t.Errorf("Failed to deserialise valid yaml:\n%s\nError: %v", input, err)
	}
	_, err = u.LatestVersion()
	if err == nil {
		t.Errorf("LatestVersion on UpstreamBase should return an error")
	}
}
