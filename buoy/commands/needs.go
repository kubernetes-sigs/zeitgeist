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

	"github.com/spf13/cobra"

	"sigs.k8s.io/zeitgeist/buoy/pkg/gomod"
)

func addNeedsCmd(root *cobra.Command) {
	var domain string

	cmd := &cobra.Command{
		Use:   "needs go.mod",
		Short: "Find dependencies based on a base import domain.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			gomods := args

			_, packages, err := gomod.Modules(gomods, domain)
			if err != nil {
				return err
			}

			for _, p := range packages {
				if p != "" {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), p)
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&domain, "domain", "d", "", "domain filter (i.e. knative.dev) [required]")

	_ = cmd.MarkFlagRequired("domain") // nolint: errcheck

	root.AddCommand(cmd)
}
