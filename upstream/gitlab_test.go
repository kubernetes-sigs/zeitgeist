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
	"time"

	"github.com/stretchr/testify/require"
	gogitlab "gitlab.com/gitlab-org/api/client-go/v2"
	"go.yaml.in/yaml/v3"
)

func TestUnserialiseGitLab(t *testing.T) {
	validYamls := []string{
		"flavour: gitlab\nurl: helm/helm\nconstraints: <1.0.0",
		"flavour: gitlab\nurl: helm/helm\nserver: https://mygitlab.com/\nconstraints: <1.0.0",
	}

	for _, valid := range validYamls {
		var u GitLab

		err := yaml.Unmarshal([]byte(valid), &u)
		require.NoError(t, err)
	}
}

func TestMostRecentGitLabRelease(t *testing.T) {
	at := func(date string) *time.Time {
		parsed, err := time.Parse(time.DateOnly, date)
		require.NoError(t, err)
		return &parsed
	}

	releases := []*gogitlab.Release{
		{TagName: "release-aaa", ReleasedAt: at("2026-09-01")},
		{TagName: "release-ccc", ReleasedAt: at("2026-10-01")},
		{TagName: "release-zzz", ReleasedAt: at("2026-08-01")},
		{TagName: "release-upcoming", ReleasedAt: at("2099-01-01"), UpcomingRelease: true},
		{TagName: "", ReleasedAt: at("2026-11-01")},
		{TagName: "release-undated"},
	}

	matchAll, err := versionFilter("")
	require.NoError(t, err)
	latest, err := mostRecentGitLabRelease(releases, matchAll)
	require.NoError(t, err)
	require.Equal(t, "release-ccc", latest)

	notCCC, err := versionFilter("^release-(aaa|zzz)$")
	require.NoError(t, err)
	latest, err = mostRecentGitLabRelease(releases, notCCC)
	require.NoError(t, err)
	require.Equal(t, "release-aaa", latest)

	matchNone, err := versionFilter("^nothing$")
	require.NoError(t, err)
	_, err = mostRecentGitLabRelease(releases, matchNone)
	require.Error(t, err)
}

func TestMostRecentGitLabReleaseNoReleases(t *testing.T) {
	matchAll, err := versionFilter("")
	require.NoError(t, err)
	_, err = mostRecentGitLabRelease(nil, matchAll)
	require.Error(t, err)
}
