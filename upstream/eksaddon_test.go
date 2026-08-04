/*
Copyright 2026 The Kubernetes Authors.

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
	"github.com/aws/aws-sdk-go-v2/service/eks"
	ekstypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"
)

type mockEKSAddonAPI func(ctx context.Context, params *eks.DescribeAddonVersionsInput, optFns ...func(*eks.Options)) (*eks.DescribeAddonVersionsOutput, error)

func (m mockEKSAddonAPI) DescribeAddonVersions(ctx context.Context, params *eks.DescribeAddonVersionsInput, optFns ...func(*eks.Options)) (*eks.DescribeAddonVersionsOutput, error) {
	return m(ctx, params, optFns...)
}

// addonVersions builds AddonVersionInfo entries with no default version marked,
// i.e. as if queried without a kubernetesVersion, or none of the versions is current.
func addonVersions(versions ...string) []ekstypes.AddonVersionInfo {
	infos := make([]ekstypes.AddonVersionInfo, 0, len(versions))
	for _, v := range versions {
		infos = append(infos, ekstypes.AddonVersionInfo{AddonVersion: aws.String(v)})
	}
	return infos
}

// addonVersionsWithDefault is like addonVersions, but marks defaultVersion as the
// one AWS reports as the "current" default for the queried Kubernetes version.
func addonVersionsWithDefault(defaultVersion string, versions ...string) []ekstypes.AddonVersionInfo {
	infos := make([]ekstypes.AddonVersionInfo, 0, len(versions))
	for _, v := range versions {
		infos = append(infos, ekstypes.AddonVersionInfo{
			AddonVersion: aws.String(v),
			Compatibilities: []ekstypes.Compatibility{
				{DefaultVersion: v == defaultVersion},
			},
		})
	}
	return infos
}

func TestEKSAddonLatestVersion(t *testing.T) {
	testCases := []struct {
		Name          string
		Input         EKSAddon
		Client        func(t *testing.T) mockEKSAddonAPI
		Expected      string
		ExpectedError string
	}{
		{
			Name: "restricts to the current (default) version by default",
			Input: EKSAddon{
				AddonName: "vpc-cni",
			},
			Client: func(_ *testing.T) mockEKSAddonAPI {
				return mockEKSAddonAPI(func(_ context.Context, _ *eks.DescribeAddonVersionsInput, _ ...func(*eks.Options)) (*eks.DescribeAddonVersionsOutput, error) {
					return &eks.DescribeAddonVersionsOutput{
						Addons: []ekstypes.AddonInfo{
							{
								AddonName: aws.String("vpc-cni"),
								// The highest version isn't the one AWS marks as current.
								AddonVersions: addonVersionsWithDefault("v1.15.1-eksbuild.1", "v1.15.1-eksbuild.1", "v1.15.4-eksbuild.2", "v1.14.0-eksbuild.3"),
							},
						},
					}, nil
				})
			},
			Expected: "v1.15.1-eksbuild.1",
		},
		{
			Name: "no default version found",
			Input: EKSAddon{
				AddonName: "vpc-cni",
			},
			Client: func(_ *testing.T) mockEKSAddonAPI {
				return mockEKSAddonAPI(func(_ context.Context, _ *eks.DescribeAddonVersionsInput, _ ...func(*eks.Options)) (*eks.DescribeAddonVersionsOutput, error) {
					return &eks.DescribeAddonVersionsOutput{
						Addons: []ekstypes.AddonInfo{
							{
								AddonName:     aws.String("vpc-cni"),
								AddonVersions: addonVersions("v1.15.1-eksbuild.1", "v1.15.4-eksbuild.2"),
							},
						},
					}, nil
				})
			},
			ExpectedError: `no default (current) version found for EKS addon "vpc-cni"`,
		},
		{
			Name: "latest: true picks the highest version regardless of the default flag",
			Input: EKSAddon{
				AddonName: "vpc-cni",
				Latest:    true,
			},
			Client: func(_ *testing.T) mockEKSAddonAPI {
				return mockEKSAddonAPI(func(_ context.Context, _ *eks.DescribeAddonVersionsInput, _ ...func(*eks.Options)) (*eks.DescribeAddonVersionsOutput, error) {
					return &eks.DescribeAddonVersionsOutput{
						Addons: []ekstypes.AddonInfo{
							{
								AddonName:     aws.String("vpc-cni"),
								AddonVersions: addonVersionsWithDefault("v1.15.1-eksbuild.1", "v1.15.1-eksbuild.1", "v1.15.4-eksbuild.2", "v1.14.0-eksbuild.3"),
							},
						},
					}, nil
				})
			},
			Expected: "v1.15.4-eksbuild.2",
		},
		{
			Name: "latest: true applies semver constraints",
			Input: EKSAddon{
				AddonName:   "vpc-cni",
				Latest:      true,
				Constraints: "< 1.15.0",
			},
			Client: func(_ *testing.T) mockEKSAddonAPI {
				return mockEKSAddonAPI(func(_ context.Context, _ *eks.DescribeAddonVersionsInput, _ ...func(*eks.Options)) (*eks.DescribeAddonVersionsOutput, error) {
					return &eks.DescribeAddonVersionsOutput{
						Addons: []ekstypes.AddonInfo{
							{
								AddonName:     aws.String("vpc-cni"),
								AddonVersions: addonVersions("v1.15.1-eksbuild.1", "v1.15.4-eksbuild.2", "v1.14.0-eksbuild.3"),
							},
						},
					}, nil
				})
			},
			Expected: "v1.14.0-eksbuild.3",
		},
		{
			Name:  "missing addon name",
			Input: EKSAddon{},
			Client: func(_ *testing.T) mockEKSAddonAPI {
				return mockEKSAddonAPI(func(_ context.Context, _ *eks.DescribeAddonVersionsInput, _ ...func(*eks.Options)) (*eks.DescribeAddonVersionsOutput, error) {
					return nil, nil
				})
			},
			ExpectedError: "EKSAddon upstream requires an addon name",
		},
		{
			Name: "API error",
			Input: EKSAddon{
				AddonName: "vpc-cni",
			},
			Client: func(_ *testing.T) mockEKSAddonAPI {
				return mockEKSAddonAPI(func(_ context.Context, _ *eks.DescribeAddonVersionsInput, _ ...func(*eks.Options)) (*eks.DescribeAddonVersionsOutput, error) {
					return nil, errors.New("AddonNotFound")
				})
			},
			ExpectedError: `retrieving EKS addon versions for "vpc-cni": AddonNotFound`,
		},
		{
			Name: "no versions found",
			Input: EKSAddon{
				AddonName: "vpc-cni",
			},
			Client: func(_ *testing.T) mockEKSAddonAPI {
				return mockEKSAddonAPI(func(_ context.Context, _ *eks.DescribeAddonVersionsInput, _ ...func(*eks.Options)) (*eks.DescribeAddonVersionsOutput, error) {
					return &eks.DescribeAddonVersionsOutput{}, nil
				})
			},
			ExpectedError: `no versions found for EKS addon "vpc-cni"`,
		},
		{
			Name: "invalid constraints",
			Input: EKSAddon{
				AddonName:   "vpc-cni",
				Constraints: "bad-constraint",
			},
			Client: func(_ *testing.T) mockEKSAddonAPI {
				return mockEKSAddonAPI(func(_ context.Context, _ *eks.DescribeAddonVersionsInput, _ ...func(*eks.Options)) (*eks.DescribeAddonVersionsOutput, error) {
					return nil, nil
				})
			},
			ExpectedError: "invalid semver constraints range: bad-constraint",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			tc.Input.ServiceClient = tc.Client(t)
			value, err := tc.Input.LatestVersion()
			if tc.ExpectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.ExpectedError)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.Expected, value)
			}
		})
	}
}

func TestEKSAddonPassesKubernetesVersion(t *testing.T) {
	var gotInput *eks.DescribeAddonVersionsInput

	e := EKSAddon{
		AddonName:         "vpc-cni",
		KubernetesVersion: "1.31",
		ServiceClient: mockEKSAddonAPI(func(_ context.Context, params *eks.DescribeAddonVersionsInput, _ ...func(*eks.Options)) (*eks.DescribeAddonVersionsOutput, error) {
			gotInput = params
			return &eks.DescribeAddonVersionsOutput{
				Addons: []ekstypes.AddonInfo{
					{AddonVersions: addonVersionsWithDefault("v1.15.1-eksbuild.1", "v1.15.1-eksbuild.1")},
				},
			}, nil
		}),
	}

	_, err := e.LatestVersion()
	require.NoError(t, err)
	require.NotNil(t, gotInput.KubernetesVersion)
	require.Equal(t, "1.31", *gotInput.KubernetesVersion)
}

func TestUnserialiseEKSAddon(t *testing.T) {
	validYamls := []string{
		"flavour: eks-addon\naddonName: vpc-cni",
		"flavour: eks-addon\naddonName: vpc-cni\nkubernetesVersion: \"1.31\"\nconstraints: < 2.0.0",
		"flavour: eks-addon\naddonName: vpc-cni\nlatest: true",
	}

	for _, valid := range validYamls {
		var u EKSAddon

		err := yaml.Unmarshal([]byte(valid), &u)
		require.NoError(t, err)
		require.NotEmpty(t, u.AddonName)
	}
}
