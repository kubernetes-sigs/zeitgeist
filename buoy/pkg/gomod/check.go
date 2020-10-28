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
	"fmt"
	"io"
	"strings"

	"github.com/blang/semver/v4"

	"sigs.k8s.io/zeitgeist/buoy/pkg/git"
	"sigs.k8s.io/zeitgeist/buoy/pkg/golang"
)

// Check examines a go mod file for dependencies and  determines if each have a release artifact
// based on the ruleset provided. Check leverages the same rules used by
// sigs.k8s.io/zeitgeist/buoy/pkg/git.Repo().BestRefFor
func Check(gomod, release, domain string, ruleset git.RulesetType, out io.Writer) error {
	modulePkgs, _, err := Modules([]string{gomod}, domain)
	if err != nil {
		return err
	}

	for module, packages := range modulePkgs {
		if err := check(module, packages, release, ruleset, out); err != nil {
			return err
		}
	}
	return nil
}

func check(module string, packages []string, release string, ruleset git.RulesetType, out io.Writer) error {
	this, err := semver.ParseTolerant(release)
	if err != nil {
		return err
	}

	if out != nil {
		_, _ = fmt.Fprintln(out, module)
	}

	nonReady := make([]string, 0)
	for _, pkg := range packages {
		repo, err := golang.ModuleToRepo(pkg)
		if err != nil {
			return err
		}

		ref, refType := repo.BestRefFor(this, ruleset)
		switch refType {
		case git.NoRef:
			nonReady = append(nonReady, ref)
			if out != nil {
				_, _ = fmt.Fprintln(out, "✘ ", ref)
			}
		default:
			if out != nil {
				_, _ = fmt.Fprintln(out, "✔ ", ref)
			}
		}
	}

	if len(nonReady) > 0 {
		return &Error{
			Module:       module,
			Dependencies: nonReady,
		}
	}

	return nil
}

// DependencyErr is a Dependency Error instance. For use with with error.Is.
var DependencyErr = &Error{}

// Error holds the result of a failed check.
type Error struct {
	Module       string
	Dependencies []string
}

var _ error = (*Error)(nil)

// Is implements error.Is(target)
func (e *Error) Is(target error) bool {
	_, is := target.(*Error)
	return is
}

// Error implements error.Error()
func (e *Error) Error() string {
	return fmt.Sprintf("%s failed because of the following dependencies [%s]",
		e.Module,
		strings.Join(e.Dependencies, ", "))
}
