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

func (m mockedReceiveMsgs) DescribeImages(in *ec2.DescribeImagesInput) (*ec2.DescribeImagesOutput, error) {
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
	require.NotEmpty(t, updates)
	require.Equal(t, updates[0].Name, "example")
	require.Equal(t, updates[0].Version, "1.0.0")
	require.Equal(t, updates[0].NewVersion, "1.0.0")
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
