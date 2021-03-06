/*
Copyright 2020 The Kubernetes Authors

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

package git

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/blang/semver/v4"
	fixtures "github.com/go-git/go-git-fixtures/v4"
)

func TestGetRepo_BasicOne(t *testing.T) {
	f := fixtures.Basic().One()
	repoURL := f.DotGit().Root()

	r, err := GetRepo("foo", repoURL)
	require.NoError(t, err)
	require.NotNil(t, r)

	// TODO: refactor test (use require pkg)
	// nolint: staticcheck
	require.Equal(t, r.DefaultBranch, "master")
	require.Len(t, r.Branches, 2)
	require.Len(t, r.Tags, 1)
}

func TestGetRepo_Error(t *testing.T) {
	_, err := GetRepo("foo", "invalid")
	require.Error(t, err)
}

func TestRepo_BestRefFor(t *testing.T) {
	repo := &Repo{
		Ref:           "ref",
		DefaultBranch: "main",
		Tags:          []string{"v0.1.0", "bar", "v0.2.0", "baz", "v0.2.1", "v0.2.2-rc.1", "v0.2.2+build", "foo"},
		Branches:      []string{"release-0.1", "bar", "release-0.2", "baz", "main", "release-0.3"},
	}

	tests := map[string]struct {
		repo    *Repo
		version semver.Version
		want    string
		release RefType
		rule    RulesetType
	}{
		"Any - v0.1": {
			repo:    repo,
			version: semver.MustParse("0.1.0"),
			want:    "ref@v0.1.0",
			release: ReleaseRef,
			rule:    AnyRule,
		},
		"Any - v0.2": {
			repo:    repo,
			version: semver.MustParse("0.2.0"),
			want:    "ref@v0.2.1",
			release: ReleaseRef,
			rule:    AnyRule,
		},
		"Any - v0.3": {
			repo:    repo,
			version: semver.MustParse("0.3.0"),
			want:    "ref@release-0.3",
			release: ReleaseBranchRef,
			rule:    AnyRule,
		},
		"Any - v0.4": {
			repo:    repo,
			version: semver.MustParse("0.4.0"),
			want:    "ref@main",
			release: DefaultBranchRef,
			rule:    AnyRule,
		},

		"ReleaseOrReleaseBranch - v0.1": {
			repo:    repo,
			version: semver.MustParse("0.1.0"),
			want:    "ref@v0.1.0",
			release: ReleaseRef,
			rule:    ReleaseOrReleaseBranchRule,
		},
		"ReleaseOrReleaseBranch - v0.2": {
			repo:    repo,
			version: semver.MustParse("0.2.0"),
			want:    "ref@v0.2.1",
			release: ReleaseRef,
			rule:    ReleaseOrReleaseBranchRule,
		},
		"ReleaseOrReleaseBranch - v0.3": {
			repo:    repo,
			version: semver.MustParse("0.3.0"),
			want:    "ref@release-0.3",
			release: ReleaseBranchRef,
			rule:    ReleaseOrReleaseBranchRule,
		},
		"ReleaseOrReleaseBranch - v0.4": {
			repo:    repo,
			version: semver.MustParse("0.4.0"),
			want:    "ref",
			release: NoRef,
			rule:    ReleaseOrReleaseBranchRule,
		},

		"Release - v0.1": {
			repo:    repo,
			version: semver.MustParse("0.1.0"),
			want:    "ref@v0.1.0",
			release: ReleaseRef,
			rule:    ReleaseRule,
		},
		"Release - v0.2": {
			repo:    repo,
			version: semver.MustParse("0.2.0"),
			want:    "ref@v0.2.1",
			release: ReleaseRef,
			rule:    ReleaseRule,
		},
		"Release - v0.3": {
			repo:    repo,
			version: semver.MustParse("0.3.0"),
			want:    "ref",
			release: NoRef,
			rule:    ReleaseRule,
		},
		"Release - v0.4": {
			repo:    repo,
			version: semver.MustParse("0.4.0"),
			want:    "ref",
			release: NoRef,
			rule:    ReleaseRule,
		},

		"ReleaseBranch - v0.1": {
			repo:    repo,
			version: semver.MustParse("0.1.0"),
			want:    "ref@release-0.1",
			release: ReleaseBranchRef,
			rule:    ReleaseBranchRule,
		},
		"ReleaseBranch - v0.2": {
			repo:    repo,
			version: semver.MustParse("0.2.0"),
			want:    "ref@release-0.2",
			release: ReleaseBranchRef,
			rule:    ReleaseBranchRule,
		},
		"ReleaseBranch - v0.3": {
			repo:    repo,
			version: semver.MustParse("0.3.0"),
			want:    "ref@release-0.3",
			release: ReleaseBranchRef,
			rule:    ReleaseBranchRule,
		},
		"ReleaseBranch - v0.4": {
			repo:    repo,
			version: semver.MustParse("0.4.0"),
			want:    "ref",
			release: NoRef,
			rule:    ReleaseBranchRule,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, release := tt.repo.BestRefFor(tt.version, tt.rule)
			require.Equal(t, got, tt.want)
			require.Equal(t, release, tt.release)
		})
	}
}

func TestNormalizeTagVersion(t *testing.T) {
	tests := map[string]struct {
		version string
		want    string
		wantOK  bool
	}{
		"v0.1.0": {
			version: "v0.1.0",
			want:    "0.1.0",
			wantOK:  true,
		},
		"v1.2.3": {
			version: "v1.2.3",
			want:    "1.2.3",
			wantOK:  true,
		},
		"notarelease": {
			version: "notarelease",
			want:    "notarelease",
			wantOK:  false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, gotOK := normalizeTagVersion(tt.version)
			require.Equal(t, gotOK, tt.wantOK)
			require.Equal(t, got, tt.want)
		})
	}
}

func TestTagVersion(t *testing.T) {
	tests := map[string]struct {
		version semver.Version
		want    string
	}{
		"v1.2.3": {
			version: semver.Version{
				Major: 1,
				Minor: 2,
				Patch: 3,
			},
			want: "v1.2.3",
		},
		"v0.1.0": {
			version: semver.Version{
				Major: 0,
				Minor: 1,
				Patch: 0,
			},
			want: "v0.1.0",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := ReleaseVersion(tt.version)
			require.Equal(t, got, tt.want)
		})
	}
}

func TestNormalizeBranchVersion(t *testing.T) {
	tests := map[string]struct {
		version string
		want    string
		wantOK  bool
	}{
		"release-0.1": {
			version: "release-0.1",
			want:    "0.1.0",
			wantOK:  true,
		},
		"release-1.2": {
			version: "release-1.2",
			want:    "1.2.0",
			wantOK:  true,
		},
		"notarelease": {
			version: "notarelease",
			want:    "notarelease",
			wantOK:  false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, gotOK := normalizeBranchVersion(tt.version)
			require.Equal(t, gotOK, tt.wantOK)
			require.Equal(t, got, tt.want)
		})
	}
}

func TestBranchVersion(t *testing.T) {
	tests := map[string]struct {
		version semver.Version
		want    string
	}{
		"v1.2.3": {
			version: semver.Version{
				Major: 1,
				Minor: 2,
				Patch: 3,
			},
			want: "release-1.2",
		},
		"v0.1.0": {
			version: semver.Version{
				Major: 0,
				Minor: 1,
				Patch: 0,
			},
			want: "release-0.1",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := ReleaseBranchVersion(tt.version)
			require.Equal(t, got, tt.want)
		})
	}
}

func TestRefType_String(t *testing.T) {
	tests := map[string]struct {
		rt   RefType
		want string
	}{
		"DefaultBranchRef": {
			rt:   DefaultBranchRef,
			want: "Default Branch",
		},
		"ReleaseBranchRef": {
			rt:   ReleaseBranchRef,
			want: "Release Branch",
		},
		"ReleaseRef": {
			rt:   ReleaseRef,
			want: "Release",
		},
		"NoRef": {
			rt:   NoRef,
			want: "No Ref",
		},
		"Invalid": {
			rt:   RefType(999),
			want: "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.rt.String()
			require.Equal(t, got, tt.want)
		})
	}
}

func TestParseRef(t *testing.T) {
	tests := []struct {
		ref         string
		wantModule  string
		wantRef     string
		wantRefType RefType
	}{
		{
			ref:         "foo@v0.1.1",
			wantModule:  "foo",
			wantRef:     "v0.1.1",
			wantRefType: ReleaseRef,
		}, {
			ref:         "foo@release-v0.1",
			wantModule:  "foo",
			wantRef:     "release-v0.1",
			wantRefType: ReleaseBranchRef,
		}, {
			ref:         "foo@default",
			wantModule:  "foo",
			wantRef:     "default",
			wantRefType: BranchRef,
		}, {
			ref:         "invalid",
			wantModule:  "invalid",
			wantRef:     "",
			wantRefType: UndefinedRef,
		}, {
			ref:         "",
			wantModule:  "",
			wantRef:     "",
			wantRefType: UndefinedRef,
		},
	}

	for _, tt := range tests {
		t.Run(tt.ref, func(t *testing.T) {
			gotModule, gotRef, gotRefType := ParseRef(tt.ref)
			require.Equal(t, gotModule, tt.wantModule)
			require.Equal(t, gotRef, tt.wantRef)
			require.Equal(t, gotRefType, tt.wantRefType)
		})
	}
}
