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
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"sigs.k8s.io/zeitgeist/buoy/pkg/gomod"
)

func addExistsCmd(root *cobra.Command) {
	var (
		release string
		verbose bool
		tag     bool
	)

	cmd := &cobra.Command{
		Use:   "exists go.mod",
		Short: "Determine if the release branch exists for a given module.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			gomodFile := args[0]

			var out io.Writer
			if verbose {
				out = cmd.OutOrStderr()
			}

			meta, err := gomod.ReleaseStatus(gomodFile, release, out)
			if err != nil {
				return err
			}

			if tag {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), meta.Release)
			}

			if !meta.ReleaseBranchExists {
				os.Exit(1)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&release, "release", "r", "", "release should be '<major>.<minor>' (i.e.: 1.23 or v1.23) [required]")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Print verbose output (stderr)")
	cmd.Flags().BoolVarP(&tag, "next", "t", false, "Print the next release tag (stdout)")

	_ = cmd.MarkFlagRequired("release") // nolint: errcheck

	root.AddCommand(cmd)
}
