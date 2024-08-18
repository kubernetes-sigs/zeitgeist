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
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/stretchr/testify/require"
)

/*
type mockedReceiveMsgs struct {
	ec2.Client
	Resp ec2.DescribeImagesOutput
}
*/

type mockEc2Api func(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error)

func (m mockEc2Api) DescribeImages(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error) {
	return m(ctx, params, optFns...)
}

func TestGetAMI(t *testing.T) {
	testCases := []struct {
		Name          string
		Input         AMI
		Resp          ec2.DescribeImagesOutput
		Client        func(t *testing.T) mockEc2Api
		Expected      string
		ExpectedError bool
	}{
		{
			Name: "AMI exist",
			Input: AMI{
				Owner: "amazon",
				Name:  "amazon-eks-node-1.13-*",
			},
			Client: func(_ *testing.T) mockEc2Api {
				return mockEc2Api(func(_ context.Context, _ *ec2.DescribeImagesInput, _ ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error) {
					return &ec2.DescribeImagesOutput{
						Images: []types.Image{
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
					}, nil
				})
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
			Client: func(_ *testing.T) mockEc2Api {
				return mockEc2Api(func(_ context.Context, _ *ec2.DescribeImagesInput, _ ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error) {
					return &ec2.DescribeImagesOutput{
						Images: []types.Image{},
					}, nil
				})
			},
			Expected:      "no AMI found for upstream this-ami-doesnt-exist-zeitgeist",
			ExpectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			tc.Input.ServiceClient = tc.Client(t)
			latestImage, err := tc.Input.LatestVersion()
			if tc.ExpectedError {
				require.Error(t, err)
				require.EqualError(t, err, tc.Expected)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.Expected, latestImage)
			}
		})
	}
}
