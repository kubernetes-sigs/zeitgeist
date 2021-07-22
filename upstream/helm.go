/*
Copyright 2020 The Kubernetes Authors.

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

package upstream

import (
	"strings"

	"github.com/blang/semver"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
)

// Helm upstream representation
type Helm struct {
	Base `mapstructure:",squash"`

	// Helm repository URL, e.g. TODO FIXME
	Repo string

	// Helm chart name in this repository
	Chart string

	// Optional: semver constraints, e.g. < 2.0.0
	// Will have no effect if the dependency does not follow Semver
	Constraints string
}

// LatestVersion returns the latest non-draft, non-prerelease Helm Release
// for the given repository (depending on the Constraints if set).
func (upstream Helm) LatestVersion() (string, error) {
	log.Debug("Using Helm flavour")
	return latestChartVersion(upstream)
}

func latestChartVersion(upstream Helm) (string, error) {
	if !strings.Contains(upstream.Repo, "//") {
		return "", errors.Errorf("invalid helm repo url: %s\nHelm repo should be a URL", upstream.Repo)
	}

	// Get the repo index first
	cfg := repo.Entry{
		Name: "zeitgeist",
		URL:  upstream.Repo,
	}
	settings := cli.EnvSettings{
		PluginsDirectory: "",
	}
	re, err := repo.NewChartRepository(&cfg, getter.All(&settings))
	if err != nil {
		return "", err
	}

	log.Debugf("Downloading repo index for %s...", upstream.Repo)
	indexFile, err := re.DownloadIndexFile()
	if err != nil {
		return "", err
	}
	index, err := repo.LoadIndexFile(indexFile)
	if err != nil {
		return "", err
	}

	log.Debugf("Loading repo index for %s...", upstream.Repo)
	chartVersions := index.Entries[upstream.Chart]
	if chartVersions == nil {
		return "", errors.Errorf("No matching chart found in this repository")
	}

	semverConstraints := upstream.Constraints
	if semverConstraints == "" {
		// If no range is passed, just use the broadest possible range
		semverConstraints = DefaultSemVerConstraints
	}

	expectedRange, err := semver.ParseRange(semverConstraints)
	if err != nil {
		return "", errors.Errorf("Invalid semver constraints range: %#v", upstream.Constraints)
	}

	// Iterate over versions and get the first newer version that matches our semver
	// Versions are already ordered, cf https://github.com/helm/helm/blob/6a3daaa7aa5b89a150042cadcbe869b477bb62a1/pkg/repo/index.go#L344
	for _, chartVersion := range chartVersions {
		version, err := semver.Parse(chartVersion.Version)
		if err != nil {
			log.Debugf("Error parsing version %s (%#v) as semver, cannot validate semver constraints", chartVersion.Version, err)
		} else if !expectedRange(version) {
			log.Debugf("Skipping release not matching range constraints (%s): %s\n", upstream.Constraints, chartVersion.Version)
			continue
		}

		log.Debugf("Found latest matching release: %s\n", chartVersion.Version)

		return chartVersion.Version, nil
	}

	//	log.Debugf("expectedRange %s", expectedRange)

	// No latest version found – no versions? Only prereleases?
	return "", errors.Errorf("No potential version found")
}
