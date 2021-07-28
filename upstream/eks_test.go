/*
Copyright 2021 The Kubernetes Authors.

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

package upstream

import (
	"testing"

	"github.com/stretchr/testify/require"

	"gopkg.in/yaml.v3"
)

func TestUnserialiseEKS(t *testing.T) {
	validYamls := []string{
		"flavour: eks",
		"flavour: eks\nconstraints: < 1.20.0",
	}

	for _, valid := range validYamls {
		var u EKS

		err := yaml.Unmarshal([]byte(valid), &u)
		require.NoError(t, err)
	}
}

func TestInvalidEKSValues(t *testing.T) {
	var err error

	badConstraint := "bad-constraint"
	e := EKS{
		Constraints: badConstraint,
	}

	_, err = e.LatestVersion()
	require.Error(t, err)
}

func TestEKSHappyPath(t *testing.T) {
	e := EKS{}

	latestVersion, err := e.LatestVersion()
	require.NoError(t, err)
	require.NotEmpty(t, latestVersion)
}

func TestEKSHappyPathWithConstraint(t *testing.T) {
	e := EKS{
		Constraints: "> 1.16.0",
	}

	latestVersion, err := e.LatestVersion()
	require.NoError(t, err)
	require.NotEmpty(t, latestVersion)
}

func TestEKSUnsatisfiableConstraint(t *testing.T) {
	e := EKS{
		Constraints: "< 1.15.0",
	}

	latestVersion, err := e.LatestVersion()
	require.Error(t, err)
	require.Empty(t, latestVersion)
}
