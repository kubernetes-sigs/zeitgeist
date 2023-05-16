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
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

type mockedReceiveMsgs struct {
	ec2iface.EC2API
	Resp ec2.DescribeImagesOutput
}

func (m mockedReceiveMsgs) DescribeImages(in *ec2.DescribeImagesInput) (*ec2.DescribeImagesOutput, error) {
	// Only need to return mocked response output
	return &m.Resp, nil
}

func TestRemoteSuccess(t *testing.T) {
	var client RemoteClient
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
	client, err := NewRemoteClient()
	require.Nil(t, err)

	_, err = client.RemoteCheck("../testdata/remote-dummy.yaml")
	require.Nil(t, err)
}

func TestDummyRemoteExportWithoutUpdate(t *testing.T) {
	client, err := NewRemoteClient()
	require.Nil(t, err)

	updates, err := client.RemoteExport("../testdata/remote-dummy.yaml")
	require.Nil(t, err)
	require.Empty(t, updates)
}

func TestDummyRemoteExportWithUpdate(t *testing.T) {
	client, err := NewRemoteClient()
	require.Nil(t, err)

	updates, err := client.RemoteExport("../testdata/remote-dummy-with-update.yaml")
	require.Nil(t, err)
	require.NotEmpty(t, updates)
	require.Equal(t, updates[0].Name, "example")
	require.Equal(t, updates[0].Version, "0.0.1")
	require.Equal(t, updates[0].NewVersion, "1.0.0")
}

func TestRemoteConstraint(t *testing.T) {
	client, err := NewRemoteClient()
	require.Nil(t, err)

	_, err = client.RemoteCheck("../testdata/remote-constraint.yaml")
	require.Nil(t, err)
}

func TestUnknownFlavour(t *testing.T) {
	client, err := NewRemoteClient()
	require.Nil(t, err)

	_, err = client.RemoteCheck("../testdata/unknown-upstream.yaml")
	require.NotNil(t, err)
}
