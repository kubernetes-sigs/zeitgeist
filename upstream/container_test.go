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
	"go.yaml.in/yaml/v3"

	"sigs.k8s.io/zeitgeist/dependency"
)

func TestUnserialiseContainer(t *testing.T) {
	validYamls := []string{
		"flavour: container\nurl: honk/honk\nconstraints: <1.0.0",
	}

	for _, valid := range validYamls {
		var u Container

		err := yaml.Unmarshal([]byte(valid), &u)
		if err != nil {
			t.Errorf("Failed to deserialise valid yaml:\n%s", valid)
		}
	}
}

func TestHighestAlphanumericTag(t *testing.T) {
	tags := []string{
		"2026-09-21-03-15-587f406",
		"2026-09-28-03-17-587f406",
		"2026-09-14-03-15-587f406",
		"2026-09-09-10-36-68cd3e0",
	}

	latest, err := highestAlphanumericVersion(tags)
	require.NoError(t, err)
	require.Equal(t, "2026-09-28-03-17-587f406", latest)

	_, err = highestAlphanumericVersion([]string{})
	require.Error(t, err)
}

func TestHighestSemverTag(t *testing.T) {
	tags := []string{"1.0.0", "v2.1.0", "2.1.0", "latest", "2026-08-01-00-00-abcdef0", "3.0.0"}

	latest, err := highestSemverTag("", tags)
	require.NoError(t, err)
	require.Equal(t, "3.0.0", latest)

	latest, err = highestSemverTag("< 3.0.0", tags)
	require.NoError(t, err)
	require.Equal(t, "2.1.0", latest)

	_, err = highestSemverTag("", []string{"main"})
	require.Error(t, err)
}

func TestContainerRandomSchemeUnsupported(t *testing.T) {
	_, err := Container{Registry: "honk/honk", Scheme: dependency.Random}.LatestVersion()
	require.ErrorIs(t, err, ErrUnsupportedScheme)
}
