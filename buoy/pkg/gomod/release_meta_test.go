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
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// TestReleaseStatus - This is an integration test, it will make a call out to the internet.
func TestReleaseStatus(t *testing.T) {
	tests := map[string]struct {
		gomod   string
		release string
		want    *ReleaseMeta
		wantErr bool
	}{
		"demo1, v0.12, knative.dev": {
			gomod:   "./testdata/gomod.next1",
			release: "v0.12",
			want: &ReleaseMeta{
				Module:              "knative.dev/serving",
				ReleaseBranchExists: true,
				ReleaseBranch:       "release-0.12",
				Release:             "v0.12.2",
			},
		},
		"demo1, v99.99, knative.dev": {
			gomod:   "./testdata/gomod.next1",
			release: "v99.88",
			want: &ReleaseMeta{
				Module:              "knative.dev/serving",
				ReleaseBranchExists: false,
				ReleaseBranch:       "release-99.88",
				Release:             "v99.88.0",
			},
		},
		"bad release": {
			gomod:   "./testdata/gomod.next1",
			release: "not gonna work",
			wantErr: true,
		},
		"bad go module": {
			gomod:   "./testdata/gomod.float1", // the module does not exist
			release: "v0.15",
			wantErr: true,
		},
		"bad go mod file": {
			gomod:   "./testdata/bad.example",
			release: "v0.15",
			wantErr: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := ReleaseStatus(tt.gomod, tt.release, os.Stdout)
			if (tt.wantErr && err == nil) || (!tt.wantErr && err != nil) {
				t.Errorf("unexpected error state, want error == %t, got %v", tt.wantErr, err)
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("Unexpected output (-got +want):\n%s", diff)
			}
		})
	}
}
