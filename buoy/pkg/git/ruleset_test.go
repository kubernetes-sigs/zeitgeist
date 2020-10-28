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

package git

import (
	"reflect"
	"testing"
)

func TestRuleset(t *testing.T) {
	tests := map[string]struct {
		rule string
		want RulesetType
	}{
		"Any": {
			rule: "Any",
			want: AnyRule,
		},
		"ReleaseOrBranch": {
			rule: "ReleaseOrBranch",
			want: ReleaseOrReleaseBranchRule,
		},
		"Release": {
			rule: "Release",
			want: ReleaseRule,
		},
		"Branch": {
			rule: "Branch",
			want: ReleaseBranchRule,
		},
		"Invalid": {
			rule: "Invalid",
			want: InvalidRule,
		},
		"Garbage": {
			rule: "dasddasdsa",
			want: InvalidRule,
		},

		"any": {
			rule: "any",
			want: AnyRule,
		},
		"releaseorbranch": {
			rule: "ReleaseOrBranch",
			want: ReleaseOrReleaseBranchRule,
		},
		"release": {
			rule: "release",
			want: ReleaseRule,
		},
		"branch": {
			rule: "Branch",
			want: ReleaseBranchRule,
		},
		"invalid": {
			rule: "invalid",
			want: InvalidRule,
		},
		"garbage": {
			rule: "adsdsaasd",
			want: InvalidRule,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := Ruleset(tt.rule); got != tt.want {
				t.Errorf("Ruleset() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRulesetType_String(t *testing.T) {
	tests := map[string]struct {
		rt   RulesetType
		want string
	}{
		"Any": {
			rt:   AnyRule,
			want: "Any",
		},
		"ReleaseOrBranch": {
			rt:   ReleaseOrReleaseBranchRule,
			want: "ReleaseOrBranch",
		},
		"Release": {
			rt:   ReleaseRule,
			want: "Release",
		},
		"Branch": {
			rt:   ReleaseBranchRule,
			want: "Branch",
		},
		"Invalid": {
			rt:   InvalidRule,
			want: "Invalid",
		},
		"Garbage": {
			rt:   RulesetType(999),
			want: "",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.rt.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRulesets(t *testing.T) {
	tests := map[string]struct {
		want []string
	}{
		"Default": {
			want: []string{"Any", "ReleaseOrBranch", "Release", "Branch"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := Rulesets(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Rulesets() = %v, want %v", got, tt.want)
			}
		})
	}
}
