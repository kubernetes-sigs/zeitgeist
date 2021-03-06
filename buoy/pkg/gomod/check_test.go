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

package gomod

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"sigs.k8s.io/zeitgeist/buoy/pkg/git"
)

// TestCheck - This is an integration test, it will make a call out to the internet.
func TestCheck(t *testing.T) {
	tests := map[string]struct {
		gomod   string
		release string
		domain  string
		rule    git.RulesetType
		wantErr bool
	}{
		"demo1, v0.15, knative.dev, any rule": {
			gomod:   "./testdata/gomod.check1",
			release: "v0.15",
			domain:  "knative.dev",
			rule:    git.AnyRule,
		},
		"demo1, v0.15, knative.dev, release rule": {
			gomod:   "./testdata/gomod.check1",
			release: "v0.15",
			domain:  "knative.dev",
			rule:    git.ReleaseRule,
			wantErr: true,
		},
		"demo1, v0.15, knative.dev, release branch rule": {
			gomod:   "./testdata/gomod.check1",
			release: "v0.15",
			domain:  "knative.dev",
			rule:    git.ReleaseBranchRule,
		},
		"demo1, v99.99, knative.dev, release branch or release rule": {
			gomod:   "./testdata/gomod.check1",
			release: "v99.99",
			domain:  "knative.dev",
			rule:    git.ReleaseOrReleaseBranchRule,
			wantErr: true,
		},
		"demo1, v0.16, knative.dev, any rule": {
			gomod:   "./testdata/gomod.check1",
			release: "v0.16",
			domain:  "knative.dev",
			rule:    git.AnyRule,
		},
		"demo1, v0.16, k8s.io, any rule": {
			gomod:   "./testdata/gomod.check1",
			release: "v0.16",
			domain:  "knative.dev",
			rule:    git.AnyRule,
		},
		"demo1, v99.99, knative.dev, any rule": {
			gomod:   "./testdata/gomod.check1",
			release: "v99.99",
			domain:  "knative.dev",
			rule:    git.AnyRule,
		},
		"bad release": {
			gomod:   "./testdata/gomod.check1",
			release: "not gonna work",
			domain:  "knative.dev",
			rule:    git.AnyRule,
			wantErr: true,
		},
		"bad go module": {
			gomod:   "./testdata/gomod.float1",
			release: "v0.15",
			domain:  "does-not-exist.nope",
			rule:    git.AnyRule,
			wantErr: true,
		},
		"bad go mod file": {
			gomod:   "./testdata/bad.example",
			release: "v0.15",
			domain:  "knative.dev",
			rule:    git.AnyRule,
			wantErr: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := Check(tt.gomod, tt.release, tt.domain, tt.rule, os.Stdout)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestError(t *testing.T) {
	tests := map[string]struct {
		err             error
		isDependencyErr bool
	}{
		"true, empty": {
			err:             &Error{},
			isDependencyErr: true,
		},
		"true, filled": {
			err: &Error{
				Module:       "foo",
				Dependencies: []string{"bar", "baz"},
			},
			isDependencyErr: true,
		},
		"false": {
			err:             errors.New("not a dep error"),
			isDependencyErr: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, errors.Is(tt.err, DependencyErr), tt.isDependencyErr)
		})
	}
}

func TestError_Error(t *testing.T) {
	tests := map[string]struct {
		err  error
		want string
	}{
		"empty": {
			err:  &Error{},
			want: " failed because of the following dependencies []",
		},
		"filled": {
			err: &Error{
				Module:       "foo",
				Dependencies: []string{"bar", "baz"},
			},
			want: "foo failed because of the following dependencies [bar, baz]",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, tt.err.Error(), tt.want)
		})
	}
}
