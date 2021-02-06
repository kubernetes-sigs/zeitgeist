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

package upstreams

import (
	"testing"

	"github.com/stretchr/testify/require"

	"gopkg.in/yaml.v3"
)

func TestUnserialiseGithub(t *testing.T) {
	validYamls := []string{
		"flavour: github\nurl: helm/helm\nconstraints: <1.0.0",
	}

	for _, valid := range validYamls {
		var u Github

		err := yaml.Unmarshal([]byte(valid), &u)
		require.NoError(t, err)
	}
}

func TestInvalidValues(t *testing.T) {
	var err error

	invalidURL := "test"
	gh := Github{
		URL: invalidURL,
	}

	_, err = gh.LatestVersion()
	require.Error(t, err)

	invalidConstraint := "invalid-constraint"
	gh2 := Github{
		URL:         "test/test",
		Constraints: invalidConstraint,
	}

	_, err = gh2.LatestVersion()
	require.Error(t, err)
}

func TestWrongRepository(t *testing.T) {
	gh := Github{
		URL: "Pluies/doesnotexist",
	}

	_, err := gh.LatestVersion()
	require.Error(t, err)
}

func TestNonExistentBranch(t *testing.T) {
	gh := Github{
		URL:    "helm/heml",
		Branch: "branch_that_does_no_exists",
	}

	_, err := gh.LatestVersion()
	if err == nil {
		t.Errorf("Failed non existent branch test. Error should not be nil")
	}
}

func TestBranchHappyPath(t *testing.T) {
	gh := Github{
		URL:    "helm/helm",
		Branch: "master",
	}

	latestVersion, err := gh.LatestVersion()
	if err != nil {
		t.Errorf("Faield github branch happy path test: %v", err)
	}

	if latestVersion == "" {
		t.Errorf("Got an empty latestVersion")
	}
}

func TestHappyPath(t *testing.T) {
	gh := Github{
		URL: "helm/helm",
	}

	latestVersion, err := gh.LatestVersion()
	require.NoError(t, err)
	require.NotEmpty(t, latestVersion)
}
