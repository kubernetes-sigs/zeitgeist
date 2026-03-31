/*
Copyright 2024 The Kubernetes Authors.

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
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type mockSSMApi func(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error)

func (m mockSSMApi) GetParameter(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
	return m(ctx, params, optFns...)
}

func TestGetSSMParameter(t *testing.T) {
	testCases := []struct {
		Name          string
		Input         SSM
		Client        func(t *testing.T) mockSSMApi
		Expected      string
		ExpectedError bool
	}{
		{
			Name: "parameter exists",
			Input: SSM{
				Name: "/aws/service/eks/optimized-ami/1.31/amazon-linux-2023/x86_64/standard/recommended/image_id",
			},
			Client: func(_ *testing.T) mockSSMApi {
				return mockSSMApi(func(_ context.Context, _ *ssm.GetParameterInput, _ ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
					return &ssm.GetParameterOutput{
						Parameter: &ssmtypes.Parameter{
							Value: aws.String("ami-1234567890abcdef0"),
						},
					}, nil
				})
			},
			Expected:      "ami-1234567890abcdef0",
			ExpectedError: false,
		},
		{
			Name: "parameter does not exist",
			Input: SSM{
				Name: "/aws/service/eks/optimized-ami/9.99/nonexistent/recommended/image_id",
			},
			Client: func(_ *testing.T) mockSSMApi {
				return mockSSMApi(func(_ context.Context, _ *ssm.GetParameterInput, _ ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
					return nil, errors.New("ParameterNotFound")
				})
			},
			Expected:      `retrieving SSM parameter "/aws/service/eks/optimized-ami/9.99/nonexistent/recommended/image_id": ParameterNotFound`,
			ExpectedError: true,
		},
		{
			Name:  "missing parameter name",
			Input: SSM{},
			Client: func(_ *testing.T) mockSSMApi {
				return mockSSMApi(func(_ context.Context, _ *ssm.GetParameterInput, _ ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
					return nil, nil
				})
			},
			Expected:      "SSM upstream requires a parameter name",
			ExpectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			tc.Input.ServiceClient = tc.Client(t)
			value, err := tc.Input.LatestVersion()
			if tc.ExpectedError {
				require.Error(t, err)
				require.EqualError(t, err, tc.Expected)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.Expected, value)
			}
		})
	}
}

func TestUnserialiseSSM(t *testing.T) {
	validYamls := []string{
		"flavour: ssm\nname: /aws/service/eks/optimized-ami/1.31/amazon-linux-2023/x86_64/standard/recommended/image_id",
		"flavour: ssm\nname: /my/custom/parameter",
	}

	for _, valid := range validYamls {
		var u SSM

		err := yaml.Unmarshal([]byte(valid), &u)
		require.NoError(t, err)
	}
}
