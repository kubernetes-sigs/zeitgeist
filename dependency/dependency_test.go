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

package dependency

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"gopkg.in/yaml.v3"
)

type mockedReceiveMsgs struct {
	ec2iface.EC2API
	Resp ec2.DescribeImagesOutput
}

func (m mockedReceiveMsgs) DescribeImages(_ *ec2.DescribeImagesInput) (*ec2.DescribeImagesOutput, error) {
	// Only need to return mocked response output
	return &m.Resp, nil
}

func TestLocalSuccess(t *testing.T) {
	client := NewClient()

	err := client.LocalCheck("../testdata/local.yaml", "../testdata")
	require.Nil(t, err)
}

func TestRemoteSuccess(t *testing.T) {
	var client Client
	client.AWSEC2Client = mockedReceiveMsgs{
		Resp: ec2.DescribeImagesOutput{
			Images: []*ec2.Image{
				{
					CreationDate: aws.String("2019-05-10T13:17:12.000Z"),
					ImageId:      aws.String("ami-09bbefc07310f7914"),
					Name:         aws.String("amazon-eks-node-1.13-honk"),
				},
			},
		},
	}

	_, err := client.RemoteCheck("../testdata/remote.yaml")
	require.Nil(t, err)
}

func TestDummyRemote(t *testing.T) {
	client := NewClient()

	_, err := client.RemoteCheck("../testdata/remote-dummy.yaml")
	require.Nil(t, err)
}

func TestDummyRemoteExportWithoutUpdate(t *testing.T) {
	client := NewClient()

	updates, err := client.RemoteExport("../testdata/remote-dummy.yaml")
	require.Nil(t, err)
	require.Empty(t, updates)
}

func TestDummyRemoteExportWithUpdate(t *testing.T) {
	client := NewClient()

	updates, err := client.RemoteExport("../testdata/remote-dummy-with-update.yaml")
	require.Nil(t, err)
	require.NotEmpty(t, updates)
	require.Equal(t, updates[0].Name, "example")
	require.Equal(t, updates[0].Version, "0.0.1")
	require.Equal(t, updates[0].NewVersion, "1.0.0")
}

func TestRemoteConstraint(t *testing.T) {
	client := NewClient()

	_, err := client.RemoteCheck("../testdata/remote-constraint.yaml")
	require.Nil(t, err)
}

func TestBrokenFile(t *testing.T) {
	client := NewClient()

	err := client.LocalCheck("../testdata/does-not-exist", "../testdata")
	require.NotNil(t, err)

	err = client.LocalCheck("../testdata/Dockerfile", "../testdata")
	require.NotNil(t, err)
}

func TestLocalOutOfSync(t *testing.T) {
	client := NewClient()

	err := client.LocalCheck("../testdata/local-out-of-sync.yaml", "../testdata")
	require.NotNil(t, err)
}

func TestLocalInvalid(t *testing.T) {
	client := NewClient()

	err := client.LocalCheck("../testdata/local-invalid.yaml", "../testdata")
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "compiling regex")
}

func TestFileDoesntExist(t *testing.T) {
	client := NewClient()

	err := client.LocalCheck("../testdata/local-no-file.yaml", "../testdata")
	require.NotNil(t, err)
}

func TestUnknownFlavour(t *testing.T) {
	client := NewClient()

	_, err := client.RemoteCheck("../testdata/unknown-upstream.yaml")
	require.NotNil(t, err)
}

func TestDeserialising(t *testing.T) {
	invalidYamls := []string{
		"a b c",
		"name:",
		"name: test",
		"version: 1.0.0",
	}

	for _, invalid := range invalidYamls {
		var d Dependency

		err := yaml.Unmarshal([]byte(invalid), &d)
		require.NotNil(t, err)
	}

	validYamls := []string{
		"name: test\nversion: 1.0.0",
		"name: test\nversion: 100",
	}

	for _, valid := range validYamls {
		var d Dependency

		err := yaml.Unmarshal([]byte(valid), &d)
		require.Nil(t, err)
	}
}

func TestCheckUpstreamVersions(t *testing.T) {
	deps := []*Dependency{
		{
			Name:        "test",
			Version:     "0.0.1",
			Scheme:      Semver,
			Sensitivity: Patch,
			Upstream: map[string]string{
				"flavour": "dummy",
			},
			RefPaths: []*RefPath{
				{
					Path:  "test",
					Match: "test",
				},
			},
		},
		{
			Name:        "test-no-upstream",
			Version:     "0.0.1",
			Scheme:      Semver,
			Sensitivity: Patch,
			RefPaths: []*RefPath{
				{
					Path:  "test",
					Match: "test",
				},
			},
		},
	}

	client := NewClient()
	updateInfos, err := client.checkUpstreamVersions(deps)
	require.Nil(t, err)

	expectedUpdateInfos := []versionUpdateInfo{
		{
			name: "test",
			current: Version{
				Version: "0.0.1",
				Scheme:  Semver,
			},
			latest: Version{
				Version: "1.0.0",
				Scheme:  Semver,
			},
			updateAvailable: true,
		},
		{
			name: "test-no-upstream",
			current: Version{
				Version: "0.0.1",
				Scheme:  Semver,
			},
			updateAvailable: false,
		},
	}

	for i, updateInfo := range updateInfos {
		if !reflect.DeepEqual(updateInfo, expectedUpdateInfos[i]) {
			t.Errorf("checkUpstreamVersions mismatch at index %d:\ngot: %#v\nexpected: %#v", i, updateInfo, expectedUpdateInfos[i])
		}
	}
}

func TestUpgrade(t *testing.T) {
	dir := t.TempDir()
	testFile := filepath.Join(dir, "test.txt")

	err := os.WriteFile(testFile, []byte("VERSION: 0.0.1\nOTHER: 0.0.1"), 0o644)
	require.Nil(t, err)

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
	require.Nil(t, err)

	client := NewClient()
	ret, err := client.Upgrade(filepath.Join(dir, "dependencies.yaml"), dir)
	if err != nil {
		t.Fatalf("Upgrade failed: %v", err)
	}

	require.Equal(t, len(ret), 1)
	require.Equal(t, ret[0], "Upgraded dependency upgrade from version 0.0.1 to version 1.0.0")

	got, err := os.ReadFile(testFile)
	require.Nil(t, err)
	require.Equal(t, string(got), "VERSION: 1.0.0\nOTHER: 0.0.1")
}
