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
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	ekstypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/blang/semver/v4"
	log "github.com/sirupsen/logrus"
)

// EKSAddon is the Elastic Kubernetes Service Add-ons upstream
//
// See: https://docs.aws.amazon.com/eks/latest/userguide/eks-add-ons.html
type EKSAddon struct {
	Base `mapstructure:",squash"`

	// The name of the add-on from the Amazon EKS API, e.g. vpc-cni, coredns, kube-proxy, aws-ebs-csi-driver
	// To retrieve the full list of addons, run:
	//   aws eks describe-addon-versions --query 'addons[].addonName'
	AddonName string `yaml:"addonName"`

	// Optional: restrict to versions compatible with this Kubernetes version, e.g. "1.31"
	KubernetesVersion string `yaml:"kubernetesVersion"`

	// Optional: semver constraints, e.g. < 2.0.0
	Constraints string

	// Optional: by default, only the version AWS marks as "current" (its recommended
	// default version) is considered. Set to true to instead consider the highest
	// available version, regardless of whether AWS has marked it as default.
	Latest bool `yaml:"latest"`

	// ServiceClient is the AWS client to talk to the EKS API
	ServiceClient EKSDescribeAddonVersionsAPI
}

type EKSDescribeAddonVersionsAPI interface {
	DescribeAddonVersions(ctx context.Context, params *eks.DescribeAddonVersionsInput, optFns ...func(*eks.Options)) (*eks.DescribeAddonVersionsOutput, error)
}

// NewEKSClient returns a new aws service client for EKS
//
// Authentication is provided by the standard AWS credentials use the standard
// `~/.aws/config` and `~/.aws/credentials` files, and support environment variables.
// See AWS documentation for more details:
// https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/configure-gosdk.html
func NewEKSClient() *eks.Client {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatal("failed to load aws config", err)
	}

	return eks.NewFromConfig(cfg)
}

// LatestVersion returns the latest available version of the EKS add-on.
func (upstream EKSAddon) LatestVersion() (string, error) {
	log.Debug("Using EKSAddon upstream")

	if upstream.AddonName == "" {
		return "", errors.New("EKSAddon upstream requires an addon name")
	}

	semverConstraints := upstream.Constraints
	if semverConstraints == "" {
		// If no range is passed, just use the broadest possible range
		semverConstraints = DefaultSemVerConstraints
	}

	expectedRange, err := semver.ParseRange(semverConstraints)
	if err != nil {
		return "", fmt.Errorf("invalid semver constraints range: %v: %w", upstream.Constraints, err)
	}

	input := &eks.DescribeAddonVersionsInput{
		AddonName: &upstream.AddonName,
	}
	if upstream.KubernetesVersion != "" {
		input.KubernetesVersion = &upstream.KubernetesVersion
	}

	result, err := upstream.ServiceClient.DescribeAddonVersions(context.Background(), input)
	if err != nil {
		return "", fmt.Errorf("retrieving EKS addon versions for %q: %w", upstream.AddonName, err)
	}

	allVersions := make([]string, 0)
	candidateVersions := make([]string, 0)
	for _, addon := range result.Addons {
		for _, addonVersion := range addon.AddonVersions {
			if addonVersion.AddonVersion == nil {
				continue
			}
			allVersions = append(allVersions, *addonVersion.AddonVersion)
			if upstream.Latest || isDefaultVersion(addonVersion) {
				candidateVersions = append(candidateVersions, *addonVersion.AddonVersion)
			}
		}
	}

	if len(allVersions) == 0 {
		return "", fmt.Errorf("no versions found for EKS addon %q", upstream.AddonName)
	}

	if len(candidateVersions) == 0 {
		return "", fmt.Errorf("no default (current) version found for EKS addon %q; set kubernetesVersion, or set latest: true to consider the highest available version instead", upstream.AddonName)
	}

	return selectHighestVersion(semverConstraints, expectedRange, candidateVersions)
}

// isDefaultVersion returns whether AWS marks this add-on version as the
// "current" default version for at least one compatible Kubernetes version.
func isDefaultVersion(addonVersion ekstypes.AddonVersionInfo) bool {
	for _, compatibility := range addonVersion.Compatibilities {
		if compatibility.DefaultVersion {
			return true
		}
	}
	return false
}
