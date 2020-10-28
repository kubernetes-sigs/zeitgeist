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

	"github.com/google/go-cmp/cmp"
)

func TestModule(t *testing.T) {
	tests := map[string]struct {
		file     string
		domain   string
		wantName string
		wantDeps []string
		wantErr  bool
	}{
		"example1, knative.dev": {
			file:     "testdata/gomod.example1",
			domain:   "knative.dev",
			wantName: "knative.dev/test-demo1",
			wantDeps: []string{"knative.dev/eventing", "knative.dev/pkg", "knative.dev/serving", "knative.dev/test-infra"},
		},
		"example1, knative.dev with extra spaces": {
			file:     "testdata/gomod.example1",
			domain:   "      knative.dev   ",
			wantName: "knative.dev/test-demo1",
			wantDeps: []string{"knative.dev/eventing", "knative.dev/pkg", "knative.dev/serving", "knative.dev/test-infra"},
		},
		"example1, k8s.io": {
			file:     "testdata/gomod.example1",
			domain:   "k8s.io",
			wantName: "knative.dev/test-demo1",
			wantDeps: []string{"k8s.io/api", "k8s.io/apimachinery", "k8s.io/client-go"},
		},
		"example1, example.com": {
			file:     "testdata/gomod.example1",
			domain:   "example.com",
			wantName: "knative.dev/test-demo1",
			wantDeps: []string{}, // non-nil empty list.
		},
		"example2, knative.dev": {
			file:     "testdata/gomod.example2",
			domain:   "knative.dev",
			wantName: "knative.dev/test-demo2",
			wantDeps: []string{"knative.dev/discovery", "knative.dev/pkg", "knative.dev/test-infra"},
		},
		"bad example": {
			file:    "testdata/bad.example",
			domain:  "knative.dev",
			wantErr: true,
		},
		"no domain": {
			file:    "testdata/bad.example",
			domain:  "knative.dev",
			wantErr: true,
		},
		"missing file": {
			file:    "does-not-exist",
			domain:  "knative.dev",
			wantErr: true,
		},
		"no file": {
			domain:  "knative.dev",
			wantErr: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			name, deps, err := Module(tt.file, tt.domain)
			if (tt.wantErr && err == nil) || (!tt.wantErr && err != nil) {
				t.Errorf("unexpected error state, want error == %t, got %v", tt.wantErr, err)
				return
			}
			if name != tt.wantName {
				t.Errorf("Module() name incorrect; got %q, want: %q", name, tt.wantName)
			}
			if diff := cmp.Diff(tt.wantDeps, deps); diff != "" {
				t.Error("Module() deps diff(-want,+got):\n", diff)
			}
		})
	}
}

func TestModules(t *testing.T) {
	tests := map[string]struct {
		files    []string
		domain   string
		wantPkgs map[string][]string
		wantDeps []string
		wantErr  bool
	}{
		"example1, example2, knative.dev": {
			files:  []string{"testdata/gomod.example1", "testdata/gomod.example2"},
			domain: "knative.dev",
			wantPkgs: map[string][]string{
				"knative.dev/test-demo1": {"knative.dev/eventing", "knative.dev/pkg", "knative.dev/serving", "knative.dev/test-infra"},
				"knative.dev/test-demo2": {"knative.dev/discovery", "knative.dev/pkg", "knative.dev/test-infra"},
			},
			wantDeps: []string{"knative.dev/discovery", "knative.dev/eventing", "knative.dev/pkg", "knative.dev/serving", "knative.dev/test-infra"},
		},
		"example1, example2, k8s.io": {
			files:  []string{"testdata/gomod.example1", "testdata/gomod.example2"},
			domain: "k8s.io",
			wantPkgs: map[string][]string{
				"knative.dev/test-demo1": {"k8s.io/api", "k8s.io/apimachinery", "k8s.io/client-go"},
				"knative.dev/test-demo2": {"k8s.io/api", "k8s.io/apimachinery", "k8s.io/client-go"},
			},
			wantDeps: []string{"k8s.io/api", "k8s.io/apimachinery", "k8s.io/client-go"},
		},
		"dup example1, knative.dev": {
			files:  []string{"testdata/gomod.example1", "testdata/gomod.example1"},
			domain: "knative.dev",
			wantPkgs: map[string][]string{
				"knative.dev/test-demo1": {"knative.dev/eventing", "knative.dev/pkg", "knative.dev/serving", "knative.dev/test-infra"},
			},
			wantDeps: []string{"knative.dev/eventing", "knative.dev/pkg", "knative.dev/serving", "knative.dev/test-infra"},
		},
		"bad example": {
			files:   []string{"testdata/gomod.example1", "testdata/bad.example"},
			domain:  "knative.dev",
			wantErr: true,
		},
		"no domain": {
			files:   []string{"testdata/gomod.example1", "testdata/gomod.example2"},
			wantErr: true,
		},
		"missing file": {
			files:   []string{"testdata/gomod.example1", "does-not-exist"},
			domain:  "knative.dev",
			wantErr: true,
		},
		"no file": {
			domain:  "knative.dev",
			wantErr: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			pkgs, deps, err := Modules(tt.files, tt.domain)
			if (tt.wantErr && err == nil) || (!tt.wantErr && err != nil) {
				t.Errorf("unexpected error state, want error == %t, got %v", tt.wantErr, err)
				return
			}
			if diff := cmp.Diff(tt.wantPkgs, pkgs); diff != "" {
				t.Error("Modules() pkgs diff(-want,+got):\n", diff)
			}

			if diff := cmp.Diff(tt.wantDeps, deps); diff != "" {
				t.Error("Modules() deps diff(-want,+got):\n", diff)
			}
		})
	}
}
