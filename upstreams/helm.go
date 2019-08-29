package upstreams

import (
	"fmt"
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"
	"k8s.io/helm/pkg/getter"
	"k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/repo"
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

// Get the latest version of a Helm chart.
//
// Returns the latest chart version in the given repository.
//
// Authentication
//
// Authentication is passed through parameters on the upstream, matching the ones you'd pass to Helm directly.
func (upstream Helm) LatestVersion() (string, error) {
	log.Debugf("Using Helm upstream")

	repoUrl := upstream.Repo
	if repoUrl == "" || repoUrl == "stable" {
		repoUrl = "https://kubernetes-charts.storage.googleapis.com/"
	}

	// Download and write the index file to a temporary location
	// TODO: cache this file to prevent unnecessary network round-trips during a single invocation
	tempIndexFile, err := ioutil.TempFile("", "tmp-repo-file")
	if err != nil {
		return "", fmt.Errorf("cannot write index file for repository requested")
	}
	defer os.Remove(tempIndexFile.Name())

	c := repo.Entry{
		URL:      repoUrl,
		Username: upstream.Username,
		Password: upstream.Password,
		CertFile: upstream.CertFile,
		KeyFile:  upstream.KeyFile,
		CAFile:   upstream.CAFile,
	}
	r, err := repo.NewChartRepository(&c, getter.All(environment.EnvSettings{}))
	if err != nil {
		return "", err
	}
	if err := r.DownloadIndexFile(tempIndexFile.Name()); err != nil {
		return "", fmt.Errorf("Looks like %q is not a valid chart repository or cannot be reached: %s", repoUrl, err)
	}

	// Read the index file for the repository to get chart information and return chart URL
	repo, err := repo.LoadIndexFile(tempIndexFile.Name())
	if err != nil {
		return "", err
	}

	cv, err := repo.Get(upstream.Name, upstream.Constraints)
	if err != nil {
		if upstream.Constraints == "" {
			return "", fmt.Errorf("%s not found in %s repository", upstream.Name, repoUrl)
		} else {
			return "", fmt.Errorf("%s not found in %s repository (with constraints: %s)", upstream.Name, repoUrl, upstream.Constraints)
		}
	}

	return cv.Version, nil
}
