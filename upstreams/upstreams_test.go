package upstreams

import (
	"testing"

	"gopkg.in/yaml.v3"
)

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
