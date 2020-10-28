/*
Copyright 2020 The Kubernetes Authors

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
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"sigs.k8s.io/zeitgeist/buoy/pkg/git"
	"sigs.k8s.io/zeitgeist/buoy/pkg/gomod"
)

func addCheckCmd(root *cobra.Command) {
	var domain string
	var release string
	var rulesetFlag string
	var ruleset git.RulesetType
	var verbose bool

	var cmd = &cobra.Command{
		Use:   "check go.mod",
		Short: "Determine if this module has a ref for each dependency for a given release based on a ruleset.",
		Long: `
The check command is used to evaluate if each dependency for the given module
meets the requirements for cutting a release branch. If the requirements are
met based on the ruleset selected, the command will exit with code 0, otherwise
an error message is generated and the with the failed dependencies and exit
code 1. Errors are written to stderr. Verbose output is written to stdout.

Rulesets,
  Release          check requires all dependencies to have tagged releases.
  Branch           check requires all dependencies to have a release branch.
  ReleaseOrBranch  check will use rule (Release || Branch).

`,
		Args: cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Validation
			ruleset = git.Ruleset(rulesetFlag)
			if ruleset == git.InvalidRule {
				return fmt.Errorf("invalid ruleset, please select one of: [%s]", strings.Join(git.Rulesets(), ", "))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			gomodFile := args[0]

			var out io.Writer
			if verbose {
				out = cmd.OutOrStderr()
			}

			err := gomod.Check(gomodFile, release, domain, ruleset, out)
			if errors.Is(err, gomod.DependencyErr) {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), err.Error())
				os.Exit(1)
			}

			return err
		},
	}

	cmd.Flags().StringVarP(&domain, "domain", "d", "", "domain filter (i.e. knative.dev) [required]")
	_ = cmd.MarkFlagRequired("domain")
	cmd.Flags().StringVarP(&release, "release", "r", "", "release should be '<major>.<minor>' (i.e.: 1.23 or v1.23) [required]")
	_ = cmd.MarkFlagRequired("release")
	cmd.Flags().StringVar(&rulesetFlag, "ruleset", git.ReleaseOrReleaseBranchRule.String(), fmt.Sprintf("The ruleset to evaluate the dependency refs. Rulesets: [%s]", strings.Join(git.Rulesets(), ", ")))
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Print verbose output.")

	root.AddCommand(cmd)
}
