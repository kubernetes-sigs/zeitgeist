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
	"testing"

	"sigs.k8s.io/zeitgeist/buoy/pkg/git"
)

// TestFloat - This is an integration test, it will make a call out to the internet.
func TestFloat(t *testing.T) {
	tests := map[string]struct {
		gomod   string
		release string
		domain  string
		rule    git.RulesetType
		want    map[string]git.RefType
	}{
		"demo1, v0.15, knative.dev, any rule": {
			gomod:   "./testdata/gomod.float1",
			release: "v0.15",
			domain:  "knative.dev",
			rule:    git.AnyRule,
			want: map[string]git.RefType{
				"knative.dev/pkg":      git.ReleaseBranchRef,
				"knative.dev/eventing": git.ReleaseRef,
			},
		},
		"demo1, v0.15, knative.dev, release rule": {
			gomod:   "./testdata/gomod.float1",
			release: "v0.15",
			domain:  "knative.dev",
			rule:    git.ReleaseRule,
			want: map[string]git.RefType{
				"knative.dev/eventing": git.ReleaseRef,
			},
		},
		"demo1, v0.15, knative.dev, release branch rule": {
			gomod:   "./testdata/gomod.float1",
			release: "v0.15",
			domain:  "knative.dev",
			rule:    git.ReleaseBranchRule,
			want: map[string]git.RefType{
				"knative.dev/pkg":      git.ReleaseBranchRef,
				"knative.dev/eventing": git.ReleaseBranchRef,
			},
		},
		"demo1, v99.99, knative.dev, release branch or release rule": {
			gomod:   "./testdata/gomod.float1",
			release: "v99.99",
			domain:  "knative.dev",
			rule:    git.ReleaseOrReleaseBranchRule,
			want:    map[string]git.RefType{},
		},
		"demo1, v0.16, knative.dev, any rule": {
			gomod:   "./testdata/gomod.float1",
			release: "v0.16",
			domain:  "knative.dev",
			rule:    git.AnyRule,
			want: map[string]git.RefType{
				"knative.dev/pkg":      git.ReleaseBranchRef,
				"knative.dev/eventing": git.ReleaseRef,
			},
		},
		"demo1, v0.16, k8s.io, any rule": {
			gomod:   "./testdata/gomod.float1",
			release: "v0.16",
			domain:  "knative.dev",
			rule:    git.AnyRule,
			want: map[string]git.RefType{
				"knative.dev/pkg":      git.ReleaseBranchRef,
				"knative.dev/eventing": git.ReleaseRef,
			},
		},
		"demo1, v99.99, knative.dev, any rule": {
			gomod:   "./testdata/gomod.float1",
			release: "v99.99",
			domain:  "knative.dev",
			rule:    git.AnyRule,
			want: map[string]git.RefType{
				"knative.dev/pkg":      git.BranchRef,
				"knative.dev/eventing": git.BranchRef,
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			deps, err := Float(tt.gomod, tt.release, tt.domain, tt.rule)
			if err != nil {
				t.Fatal(err)
			}
			for _, dep := range deps {
				module, _, got := git.ParseRef(dep)
				if want, ok := tt.want[module]; ok {
					if got != want {
						t.Errorf("Float() %s; got %q, want: %q", module, got, want)
					}
				} else {
					t.Error("untested float dep: ", dep)
				}
			}
		})
	}
}

// TestFloat_unhappy - This is an integration test, it will make a call out to the internet.
func TestFloat_unhappy(t *testing.T) {
	tests := map[string]struct {
		gomod   string
		release string
		domain  string
		rule    git.RulesetType
	}{
		"bad go mod file": {
			gomod:   "./testdata/bad.example",
			release: "v0.15",
			domain:  "knative.dev",
			rule:    git.AnyRule,
		},
		"bad go module": {
			gomod:   "./testdata/gomod.float1",
			release: "v0.15",
			domain:  "does-not-exist.nope",
			rule:    git.AnyRule,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := Float(tt.gomod, tt.release, tt.domain, tt.rule)
			if err == nil {
				t.Errorf("expected to error")
			}
		})
	}
}
