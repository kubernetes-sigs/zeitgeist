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

package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"sigs.k8s.io/zeitgeist/dependencies"
)

type ValidateOptions struct {
	local  bool
	remote bool
	config string
}

var validateOpts = &ValidateOptions{}

const defaultConfigFile = "dependencies.yaml"

// validateCmd is a subcommand which invokes RunValidate()
var validateCmd = &cobra.Command{
	Use:           "validate",
	Short:         "Check dependencies locally and against upstream versions",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return RunValidate(validateOpts)
	},
}

func init() {
	// Submit types
	validateCmd.PersistentFlags().BoolVar(
		&validateOpts.local,
		"local",
		false,
		"validate dependencies locally",
	)

	validateCmd.PersistentFlags().BoolVar(
		&validateOpts.remote,
		"remote",
		false,
		"validate dependencies against specified upstreams",
	)

	validateCmd.PersistentFlags().StringVar(
		&validateOpts.config,
		"config",
		defaultConfigFile,
		"location of zeitgeist configuration file",
	)

	rootCmd.AddCommand(validateCmd)
}

// RunValidate is the function invoked by 'krel gcbmgr', responsible for
// submitting release jobs to GCB
func RunValidate(opts *ValidateOptions) error {
	if opts.remote {
		updates, err := dependencies.RemoteCheck(opts.config)
		if err != nil {
			return errors.Wrap(err, "check remote dependencies")
		}

		for _, update := range updates {
			fmt.Println(update)
		}
	}

	return dependencies.LocalCheck(opts.config)
}
