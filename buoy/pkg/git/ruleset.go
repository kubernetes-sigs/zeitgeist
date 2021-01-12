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

import "strings"

// RulesetType defines the rules to use for calculating repo.BestRefFor.
type RulesetType int

const (
	// AnyRule - release tag, release branch, or default branch
	AnyRule RulesetType = iota
	// ReleaseOrReleaseBranchRule - only release tag or release branch
	ReleaseOrReleaseBranchRule
	// ReleaseRule - only release tag
	ReleaseRule
	// ReleaseBranchRule - only release branch
	ReleaseBranchRule
	// InvalidRule - unable to parse
	InvalidRule
)

var (
	rulesetTypeString = []string{"Any", "ReleaseOrBranch", "Release", "Branch", "Invalid"}
	rulesetLookup     map[string]RulesetType
)

// init will produce a ruleset lookup map to help with ruleset string conversion.
func init() {
	rulesetLookup = make(map[string]RulesetType, len(rulesetTypeString))
	for i, rt := range rulesetTypeString {
		rule := RulesetType(i)
		rulesetLookup[strings.ToLower(rt)] = rule
	}
}

// String returns the string represented by the Ruleset.
func (rt RulesetType) String() string {
	if rt >= AnyRule && rt <= InvalidRule {
		return rulesetTypeString[rt]
	}
	return ""
}

// Ruleset converts a rule string into a RulesetType.
func Ruleset(rule string) RulesetType {
	if r, found := rulesetLookup[strings.ToLower(rule)]; found {
		return r
	}
	return InvalidRule
}

// Rulesets returns the valid strings to use to parse into a RulesetType.
func Rulesets() []string {
	return []string{
		AnyRule.String(),
		ReleaseOrReleaseBranchRule.String(),
		ReleaseRule.String(),
		ReleaseBranchRule.String(),
		// Invalid is omitted.
	}
}
