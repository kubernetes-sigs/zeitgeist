/*
Copyright 2023 The Kubernetes Authors.

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

package dependency

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/stretchr/testify/require"

	deppkg "sigs.k8s.io/zeitgeist/dependency"
)

type mockedEc2DescribeImagesAPI struct {
	Resp ec2.DescribeImagesOutput
}

func (m mockedEc2DescribeImagesAPI) DescribeImages(_ context.Context, _ *ec2.DescribeImagesInput, _ ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error) {
	// Only need to return mocked response output
	return &m.Resp, nil
}

func TestRemoteSuccess(t *testing.T) {
	var client RemoteClient
	client.AWSEC2Client = mockedEc2DescribeImagesAPI{
		Resp: ec2.DescribeImagesOutput{
			Images: []types.Image{
				{
					CreationDate: aws.String("2019-05-10T13:17:12.000Z"),
					ImageId:      aws.String("ami-09bbefc07310f7914"),
					Name:         aws.String("amazon-eks-node-1.13-honk"),
				},
			},
		},
	}

	_, err := client.RemoteCheck("../testdata/remote.yaml")
	require.NoError(t, err)
}

func TestDummyRemote(t *testing.T) {
	client, err := NewRemoteClient(true)
	require.NoError(t, err)

	_, err = client.RemoteCheck("../testdata/remote-dummy.yaml")
	require.NoError(t, err)
}

func TestDummyRemoteExportWithoutUpdate(t *testing.T) {
	client, err := NewRemoteClient(true)
	require.NoError(t, err)

	updates, err := client.RemoteExport("../testdata/remote-dummy.yaml")
	require.NoError(t, err)
	require.Empty(t, updates)
}

func TestDummyRemoteExportWithUpdate(t *testing.T) {
	client, err := NewRemoteClient(true)
	require.NoError(t, err)

	updates, err := client.RemoteExport("../testdata/remote-dummy-with-update.yaml")
	require.NoError(t, err)
	require.NotEmpty(t, updates)
	require.Equal(t, "example", updates[0].Name)
	require.Equal(t, "0.0.1", updates[0].Version)
	require.Equal(t, "1.0.0", updates[0].NewVersion)
}

func TestRemoteConstraint(t *testing.T) {
	client, err := NewRemoteClient(true)
	require.NoError(t, err)

	_, err = client.RemoteCheck("../testdata/remote-constraint.yaml")
	require.NoError(t, err)
}

func TestUnknownFlavour(t *testing.T) {
	client, err := NewRemoteClient(true)
	require.NoError(t, err)

	_, err = client.RemoteCheck("../testdata/unknown-upstream.yaml")
	require.Error(t, err)
}

func TestCheckUpstreamVersions(t *testing.T) {
	deps := []*deppkg.Dependency{
		{
			Name:        "test",
			Version:     "0.0.1",
			Scheme:      deppkg.Semver,
			Sensitivity: deppkg.Patch,
			Upstream: map[string]string{
				"flavour": "dummy",
			},
			RefPaths: []*deppkg.RefPath{
				{
					Path:  "test",
					Match: "test",
				},
			},
		},
		{
			Name:        "test-no-upstream",
			Version:     "0.0.1",
			Scheme:      deppkg.Semver,
			Sensitivity: deppkg.Patch,
			RefPaths: []*deppkg.RefPath{
				{
					Path:  "test",
					Match: "test",
				},
			},
		},
	}

	client, err := NewRemoteClient(true)
	require.NoError(t, err)
	updateInfos, err := client.CheckUpstreamVersions(deps)
	require.NoError(t, err)

	expectedUpdateInfos := []deppkg.VersionUpdateInfo{
		{
			Name: "test",
			Current: deppkg.Version{
				Version: "0.0.1",
				Scheme:  deppkg.Semver,
			},
			Latest: deppkg.Version{
				Version: "1.0.0",
				Scheme:  deppkg.Semver,
			},
			UpdateAvailable: true,
		},
		{
			Name: "test-no-upstream",
			Current: deppkg.Version{
				Version: "0.0.1",
				Scheme:  deppkg.Semver,
			},
			UpdateAvailable: false,
		},
	}

	for i, updateInfo := range updateInfos {
		if !reflect.DeepEqual(updateInfo, expectedUpdateInfos[i]) {
			t.Errorf("checkUpstreamVersions mismatch at index %d:\ngot: %#v\nexpected: %#v", i, updateInfo, expectedUpdateInfos[i])
		}
	}
}

func TestCheckUpstreamVersionsTolerant(t *testing.T) {
	deps := []*deppkg.Dependency{
		{
			Name:        "test",
			Version:     "0.0.1",
			Scheme:      deppkg.Semver,
			Sensitivity: deppkg.Patch,
			Upstream: map[string]string{
				"flavour": "dummy",
				"latest":  "2.0",
			},
			RefPaths: []*deppkg.RefPath{
				{
					Path:  "test",
					Match: "test",
				},
			},
		},
	}

	client, err := NewRemoteClient(true)
	require.NoError(t, err)
	updateInfos, err := client.CheckUpstreamVersions(deps)
	require.NoError(t, err)

	expectedUpdateInfos := []deppkg.VersionUpdateInfo{
		{
			Name: "test",
			Current: deppkg.Version{
				Version: "0.0.1",
				Scheme:  deppkg.Semver,
			},
			Latest: deppkg.Version{
				Version: "2.0",
				Scheme:  deppkg.Semver,
			},
			UpdateAvailable: true,
		},
	}

	for i, updateInfo := range updateInfos {
		if !reflect.DeepEqual(updateInfo, expectedUpdateInfos[i]) {
			t.Errorf("checkUpstreamVersions mismatch at index %d:\ngot: %#v\nexpected: %#v", i, updateInfo, expectedUpdateInfos[i])
		}
	}
}

func TestFormatVersion(t *testing.T) {
	tests := []struct {
		name     string
		template string
		version  string
		want     string
	}{
		{
			name:     "Both versions start with 'v'",
			template: "v1.0.0",
			version:  "v2.0.0",
			want:     "v2.0.0",
		},
		{
			name:     "Template starts with 'v', version does not",
			template: "v1.0.0",
			version:  "2.0.0",
			want:     "v2.0.0",
		},
		{
			name:     "Template does not start with 'v', version does",
			template: "1.0.0",
			version:  "v2.0.0",
			want:     "2.0.0",
		},
		{
			name:     "Neither version starts with 'v'",
			template: "1.0.0",
			version:  "2.0.0",
			want:     "2.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatVersion(tt.template, tt.version); got != tt.want {
				t.Errorf("formatVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckUpstreamVersionsPreservesVersionFormat(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		upstreamLatest string
		expectedLatest string
	}{
		{
			name:           "no-v current, v-prefixed upstream",
			currentVersion: "0.0.1",
			upstreamLatest: "v1.0.0",
			expectedLatest: "1.0.0",
		},
		{
			name:           "v-prefixed current, no-v upstream",
			currentVersion: "v0.0.1",
			upstreamLatest: "1.0.0",
			expectedLatest: "v1.0.0",
		},
		{
			name:           "both v-prefixed",
			currentVersion: "v0.0.1",
			upstreamLatest: "v1.0.0",
			expectedLatest: "v1.0.0",
		},
		{
			name:           "neither v-prefixed",
			currentVersion: "0.0.1",
			upstreamLatest: "1.0.0",
			expectedLatest: "1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps := []*deppkg.Dependency{
				{
					Name:        "test",
					Version:     tt.currentVersion,
					Scheme:      deppkg.Semver,
					Sensitivity: deppkg.Patch,
					Upstream: map[string]string{
						"flavour": "dummy",
						"latest":  tt.upstreamLatest,
					},
					RefPaths: []*deppkg.RefPath{
						{Path: "test", Match: "test"},
					},
				},
			}

			client, err := NewRemoteClient()
			require.NoError(t, err)
			updateInfos, err := client.CheckUpstreamVersions(deps)
			require.NoError(t, err)
			require.Len(t, updateInfos, 1)
			require.Equal(t, tt.expectedLatest, updateInfos[0].Latest.Version)
		})
	}
}

func TestUpgradePreservesVersionFormat(t *testing.T) {
	dir := t.TempDir()
	testFile := filepath.Join(dir, "test.txt")

	err := os.WriteFile(testFile, []byte("VERSION_NO_V: 0.0.1\nVERSION_V: v0.0.1"), 0o644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(dir, "dependencies.yaml"), []byte(`
dependencies:
  - name: no-v-current
    version: 0.0.1
    scheme: semver
    upstream:
      flavour: dummy
      latest: v1.0.0
    refPaths:
    - path: test.txt
      match: VERSION_NO_V
  - name: v-current
    version: v0.0.1
    scheme: semver
    upstream:
      flavour: dummy
      latest: 1.0.0
    refPaths:
    - path: test.txt
      match: VERSION_V
`), 0o644)
	require.NoError(t, err)

	client, err := NewRemoteClient()
	require.NoError(t, err)
	ret, err := client.Upgrade(filepath.Join(dir, "dependencies.yaml"), dir)
	require.NoError(t, err)
	require.Len(t, ret, 2)

	got, err := os.ReadFile(testFile)
	require.NoError(t, err)
	// no-v current should strip v from upstream; v-current should add v to upstream
	require.Equal(t, "VERSION_NO_V: 1.0.0\nVERSION_V: v1.0.0", string(got))
}

func TestUpgrade(t *testing.T) {
	dir := t.TempDir()
	testFile := filepath.Join(dir, "test.txt")

	err := os.WriteFile(testFile, []byte("VERSION: 0.0.1\nOTHER: 0.0.1"), 0o644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(dir, "dependencies.yaml"), []byte(`
dependencies:
  - name: upgrade
    version: 0.0.1
    scheme: semver
    upstream:
      flavour: dummy
      url: example/example
    refPaths:
    - path: test.txt
      match: VERSION
  - name: no-upstream
    version: 0.0.1
    scheme: semver
    refPaths:
    - path: test.txt
      match: OTHER
`), 0o644)
	require.NoError(t, err)

	client, err := NewRemoteClient(true)
	require.NoError(t, err)
	ret, err := client.Upgrade(filepath.Join(dir, "dependencies.yaml"), dir)
	if err != nil {
		t.Fatalf("Upgrade failed: %v", err)
	}

	require.Len(t, ret, 1)
	require.Equal(t, "Upgraded dependency upgrade from version 0.0.1 to version 1.0.0", ret[0])

	got, err := os.ReadFile(testFile)
	require.NoError(t, err)
	require.Equal(t, "VERSION: 1.0.0\nOTHER: 0.0.1", string(got))
}
