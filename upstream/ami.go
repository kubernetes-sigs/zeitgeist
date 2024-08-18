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
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	log "github.com/sirupsen/logrus"
)

// AMI is the Amazon Machine Image upstream
//
// See: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/AMIs.html
type AMI struct {
	Base `mapstructure:",squash"`

	// Either owner alias (e.g. "amazon") or owner id
	Owner string

	// Name predicate, as used in --filter
	// Supports wilcards
	Name string

	// ServiceClient is the AWS client to talk to AWS API
	ServiceClient EC2DescribeImagesAPI
}

type EC2DescribeImagesAPI interface {
	DescribeImages(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error)
}

// NewAWSClient return a new aws service client for ec2
//
// Authentication is provided by the standard AWS credentials use the standard
// `~/.aws/config` and `~/.aws/credentials` files, and support environment variables.
// See AWS documentation for more details:
// https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/sessions.html
func NewAWSClient() *ec2.Client {
	// Create a new session based on shared / env credentials
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal("failed to load aws config", err)
	}

	return ec2.NewFromConfig(cfg)
}

// LatestVersion returns the latest version of an AMI.
//
// Returns the latest ami id (e.g. `ami-1234567`) from all AMIs matching the predicates, sorted by CreationDate.
//
// If images cannot be listed, or if no image matches the predicates, it will return an error instead.
func (upstream AMI) LatestVersion() (string, error) {
	log.Debug("Using AMI upstream")

	// Generate filters based on configuration
	var filters []types.Filter
	filters = append(filters, types.Filter{
		Name:   aws.String("name"),
		Values: []string{upstream.Name},
	})

	input := &ec2.DescribeImagesInput{
		Owners:  []string{upstream.Owner},
		Filters: filters,
	}

	// Do the actual API call
	result, err := upstream.ServiceClient.DescribeImages(context.TODO(), input)
	if err != nil {
		return "", err
	}

	images := result.Images

	// Sort images by creation time, so we can return the latest
	sort.Slice(images, func(i, j int) bool { return *images[i].CreationDate > *images[j].CreationDate })
	log.Debugf("Matched AMIs:\n%v", images)

	if len(images) < 1 {
		return "", fmt.Errorf("no AMI found for upstream %s", upstream.Name)
	}

	latestImage := images[0]
	log.Debugf("Latest AMI ID: %v\n", *latestImage.ImageId)

	return *latestImage.ImageId, nil
}
