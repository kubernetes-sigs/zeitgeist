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

package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"sigs.k8s.io/release-utils/log"
)

const defaultConfigFile = "dependencies.yaml"

var (
	rootOpts = &options{}

	// TODO: Implement these as a separate function or subcommand to avoid the
	//       deadcode,unused,varcheck nolints
	// Variables set by GoReleaser on release
	version = "dev"     // nolint: deadcode,unused,varcheck
	commit  = "none"    // nolint: deadcode,unused,varcheck
	date    = "unknown" // nolint: deadcode,unused,varcheck
)

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "zeitgeist",
		Short:             "Zeitgeist is a language-agnostic dependency checker",
		PersistentPreRunE: initLogging,
	}

	// Submit types
	cmd.PersistentFlags().BoolVar(
		&rootOpts.local,
		"local",
		false,
		"if specified, subcommands will only perform local checks",
	)

	cmd.PersistentFlags().BoolVar(
		&rootOpts.remote,
		"remote",
		false,
		"if specified, subcommands will query against remotes defined in the config",
	)

	cmd.PersistentFlags().StringVar(
		&rootOpts.configFile,
		"config",
		defaultConfigFile,
		"configuration file location",
	)

	cmd.PersistentFlags().StringVar(
		&rootOpts.basePath,
		"base-path",
		"",
		"base path to begin searching for dependencies (defaults to where the program was called from)",
	)

	cmd.PersistentFlags().StringVar(
		&rootOpts.logLevel,
		"log-level",
		"info",
		fmt.Sprintf("the logging verbosity, either %s", log.LevelNames()),
	)

	AddCommands(cmd)
	return cmd
}

func AddCommands(topLevel *cobra.Command) {
	addValidate(topLevel)
}

func initLogging(*cobra.Command, []string) error {
	return log.SetupGlobalLogger(rootOpts.logLevel)
}
