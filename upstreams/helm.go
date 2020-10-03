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

package upstreams

import (
	"os"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
)

// Helm upstream
type Helm struct {
	UpstreamBase `mapstructure:",squash"`
	// URL of the repository
	// If left blank, defaults to "stable", i.e. https://kubernetes-charts.storage.googleapis.com/
	Repo string
	// Name of the Helm chart
	Name string
	// Optional: semver constraints, e.g. < 2.0.0
	// Will have no effect if the dependency does not follow Semver
	Constraints string
	// Optional: authentication options
	Username string
	Password string
	CertFile string
	KeyFile  string
	CAFile   string
}

const defaultRepoURL = "https://kubernetes-charts.storage.googleapis.com/"

// TODO: Does this need to be a global variable?
// Cache remote repositories locally to prevent unnecessary network round-trips
// nolint: gochecknoglobals
var cache map[string]*repo.IndexFile

// getIndex returns the index for the given repository, and caches it for subsequent calls
func getIndex(c *repo.Entry) (*repo.IndexFile, error) {
	// Check cache first
	if cache == nil {
		// No cache: initialise it
		cache = make(map[string]*repo.IndexFile)
	} else {
		index, cacheHit := cache[c.URL]
		if cacheHit {
			log.Debugf("Using cached index for %s", c.URL)
			return index, nil
		}
	}

	envSettings := cli.EnvSettings{}
	chartRepo, err := repo.NewChartRepository(c, getter.All(&envSettings))
	if err != nil {
		return nil, err
	}

	indexFilePath, err := chartRepo.DownloadIndexFile()
	if err != nil {
		return nil, errors.Errorf("looks like %q is not a valid chart repository or cannot be reached: %s", c.URL, err)
	}

	defer os.Remove(indexFilePath)

	index, err := repo.LoadIndexFile(indexFilePath)
	if err != nil {
		return nil, err
	}

	// Found: add to cache
	cache[c.URL] = index

	return index, nil
}

// LatestVersion returns the latest version of a Helm chart.
//
// Returns the latest chart version in the given repository.
//
// Authentication
//
// Authentication is passed through parameters on the upstream, matching the ones you'd pass to Helm directly.
func (upstream *Helm) LatestVersion() (string, error) {
	log.Debugf("Using Helm upstream")

	repoURL := upstream.Repo
	if repoURL == "" || repoURL == "stable" {
		repoURL = defaultRepoURL
	}

	entry := repo.Entry{
		URL:      repoURL,
		Username: upstream.Username,
		Password: upstream.Password,
		CertFile: upstream.CertFile,
		KeyFile:  upstream.KeyFile,
		CAFile:   upstream.CAFile,
	}

	// Get the index
	index, err := getIndex(&entry)
	if err != nil {
		return "", err
	}

	cv, err := index.Get(upstream.Name, upstream.Constraints)
	if err != nil {
		if upstream.Constraints != "" {
			return "", errors.Errorf(
				"%s not found in %s repository (with constraints: %s)",
				upstream.Name,
				repoURL,
				upstream.Constraints,
			)
		}

		return "", errors.Errorf("%s not found in %s repository", upstream.Name, repoURL)
	}

	return cv.Version, nil
}
