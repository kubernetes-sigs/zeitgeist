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

	"github.com/blang/semver/v4"

	"sigs.k8s.io/zeitgeist/buoy/pkg/git"
	"sigs.k8s.io/zeitgeist/buoy/pkg/golang"
)

// ReleaseMeta holds metadata important to module release status.
type ReleaseMeta struct {
	Module              string
	ReleaseBranchExists bool
	ReleaseBranch       string
	Release             string
}

// ReleaseStatus collects metadata about release branch status and next released
// version tags for a given module.
func ReleaseStatus(gomod, release string, out io.Writer) (*ReleaseMeta, error) {
	this, err := semver.ParseTolerant(release)
	if err != nil {
		return nil, err
	}

	module, _, err := Module(gomod, "domain filter ignored")
	if err != nil {
		return nil, err
	}

	if out != nil {
		_, _ = fmt.Fprintln(out, module)
	}

	next := &ReleaseMeta{Module: module}

	repo, err := golang.ModuleToRepo(module)
	if err != nil {
		return nil, err
	}

	ref, refType := repo.BestRefFor(this, git.ReleaseBranchRule)
	if refType == git.ReleaseBranchRef {
		_, rb, _ := git.ParseRef(ref)
		next.ReleaseBranch = rb
		next.ReleaseBranchExists = true

		if out != nil {
			_, _ = fmt.Fprintln(out, "✔ ", rb)
		}
	} else {
		next.ReleaseBranch = git.ReleaseBranchVersion(this)
		next.ReleaseBranchExists = false

		if out != nil {
			_, _ = fmt.Fprintln(out, "✘ ", next.ReleaseBranch)
		}
	}

	ref, refType = repo.BestRefFor(this, git.ReleaseRule)
	if refType == git.ReleaseRef {
		_, r, _ := git.ParseRef(ref)

		// TODO: refactor to check the error
		// nolint: errcheck
		rv, _ := semver.ParseTolerant(r) // has to parse, r is from BestRefFor
		rv.Patch++

		next.Release = git.ReleaseVersion(rv)
	} else {
		next.Release = git.ReleaseVersion(this)
	}

	if out != nil {
		_, _ = fmt.Fprintln(out, "➜ ", next.Release)
	}

	return next, nil
}
