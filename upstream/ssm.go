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
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	log "github.com/sirupsen/logrus"
)

// SSM is the AWS Systems Manager Parameter Store upstream
//
// Retrieves a value stored in SSM Parameter Store, e.g. EKS recommended AMI IDs.
// See: https://docs.aws.amazon.com/eks/latest/userguide/retrieve-ami-id.html
type SSM struct {
	Base `mapstructure:",squash"`

	// The SSM parameter path, e.g.:
	// /aws/service/eks/optimized-ami/1.31/amazon-linux-2023/x86_64/standard/recommended/image_id
	Path string

	// AWS SSM client used to retrieve the parameter
	ServiceClient SSMGetParameterAPI
}

type SSMGetParameterAPI interface {
	GetParameter(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error)
}

// NewSSMClient returns a new aws service client for SSM Parameter Store
//
// Authentication is provided by the standard AWS credentials use the standard
// `~/.aws/config` and `~/.aws/credentials` files, and support environment variables.
// See AWS documentation for more details:
// https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/sessions.html
func NewSSMClient() *ssm.Client {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatal("failed to load aws config", err)
	}

	return ssm.NewFromConfig(cfg)
}

// LatestVersion returns the value of the SSM parameter as the latest version.
func (upstream SSM) LatestVersion() (string, error) {
	log.Debug("Using SSM upstream")

	if upstream.Path == "" {
		return "", fmt.Errorf("SSM upstream requires a parameter path")
	}

	input := &ssm.GetParameterInput{
		Name: &upstream.Path,
	}

	result, err := upstream.ServiceClient.GetParameter(context.Background(), input)
	if err != nil {
		return "", fmt.Errorf("retrieving SSM parameter %q: %w", upstream.Path, err)
	}

	if result.Parameter == nil || result.Parameter.Value == nil {
		return "", fmt.Errorf("SSM parameter %q has no value", upstream.Path)
	}

	value := *result.Parameter.Value
	log.Debugf("SSM parameter %q value: %s", upstream.Path, value)

	return value, nil
}
