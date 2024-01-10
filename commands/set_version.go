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

package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"sigs.k8s.io/zeitgeist/dependency"
)

func addSetVersion(topLevel *cobra.Command) {
	vo := rootOpts

	cmd := &cobra.Command{
		Use:           "set-version <dependency> <version>",
		Short:         "Set version of dependency based on given input",
		SilenceUsage:  true,
		SilenceErrors: true,
		PreRunE: func(*cobra.Command, []string) error {
			return vo.setAndValidate()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetVersion(vo, args)
		},
	}

	topLevel.AddCommand(cmd)
}

// runSetVersion is the function invoked by 'addSetVersion', responsible for
// upgrading/downgrading a single dependency to the specified version.
func runSetVersion(opts *options, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("expected exactly two arguments: <dependency> <version>")
	}

	client := dependency.NewClient()

	// Check locally first: it's fast, and ensures we're working on clean files
	if err := client.LocalCheck(opts.configFile, opts.basePath); err != nil {
		return fmt.Errorf("checking local dependencies: %w", err)
	}

	dependency, version := args[0], args[1]
	if err := client.SetVersion(opts.configFile, opts.basePath, dependency, version); err != nil {
		return fmt.Errorf("set dependency version: %w", err)
	}

	return nil
}
