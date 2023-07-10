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

package upstream

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

func (m mockedReceiveMsgs) DescribeImages(_ *ec2.DescribeImagesInput) (*ec2.DescribeImagesOutput, error) {
	// Only need to return mocked response output
	return &m.Resp, nil
}

func TestGetAMI(t *testing.T) {
	testCases := []struct {
		Name          string
		Input         AMI
		Resp          ec2.DescribeImagesOutput
		Expected      string
		ExpectedError bool
	}{
		{
			Name: "AMI exist",
			Input: AMI{
				Owner: "amazon",
				Name:  "amazon-eks-node-1.13-*",
			},
			Resp: ec2.DescribeImagesOutput{
				Images: []*ec2.Image{
					{
						CreationDate: aws.String("2019-05-10T13:17:12.000Z"),
						ImageId:      aws.String("ami-123oldimage"),
						Name:         aws.String("amazon-eks-node-1.13-honk"),
					},
					{
						CreationDate: aws.String("2019-05-12T13:17:12.000Z"),
						ImageId:      aws.String("ami-honk"),
						Name:         aws.String("amazon-eks-node-1.13-old"),
					},
				},
			},
			Expected:      "ami-honk",
			ExpectedError: false,
		},
		{
			Name: "AMI does not exist",
			Input: AMI{
				Owner: "honk",
				Name:  "this-ami-doesnt-exist-zeitgeist",
			},
			Resp: ec2.DescribeImagesOutput{
				Images: []*ec2.Image{},
			},
			Expected:      "no AMI found for upstream this-ami-doesnt-exist-zeitgeist",
			ExpectedError: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			tc.Input.ServiceClient = mockedReceiveMsgs{Resp: tc.Resp}

			latestImage, err := tc.Input.LatestVersion()
			if tc.ExpectedError {
				require.NotNil(t, err)
				require.EqualError(t, err, tc.Expected)
			} else {
				require.Nil(t, err)
				require.Equal(t, tc.Expected, latestImage)
			}
		})
	}
}
