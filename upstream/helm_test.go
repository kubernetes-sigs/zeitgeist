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
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"gopkg.in/yaml.v3"
)

func TestUnserialiseHelm(t *testing.T) {
	validYamls := []string{
		"flavour: helm\nrepo: http://example.com/repo\nchart: example",
		"flavour: helm\nrepo: https://example.com/repo\nchart: example",
		"flavour: helm\nrepo: https://example.com/repo\nchart: example\nconstraints: < 1.0.0",
	}

	for _, valid := range validYamls {
		var u Helm

		err := yaml.Unmarshal([]byte(valid), &u)
		require.NoError(t, err)
	}
}

func TestInvalidHelmValues(t *testing.T) {
	var err error

	h0 := Helm{
		// Missing repo
		Chart: "example",
	}

	_, err = h0.LatestVersion()
	require.Error(t, err)

	invalidURL := "not-a-repo"
	h1 := Helm{
		Repo:  invalidURL,
		Chart: "example",
	}

	_, err = h1.LatestVersion()
	require.Error(t, err)

	invalidConstraint := "invalid-constraint"
	h2 := Helm{
		Repo:        "http://example.com/test",
		Chart:       "example",
		Constraints: invalidConstraint,
	}

	_, err = h2.LatestVersion()
	require.Error(t, err)

	emptyChartName := ""
	h3 := Helm{
		Repo:  "http://example.com/test",
		Chart: emptyChartName,
	}

	_, err = h3.LatestVersion()
	require.Error(t, err)
}

// Now onto tests that connect to a "real" Helm repo!

// Set up a local webserver to serve the Helm repo index
// (As far as I can tell, it can't read directly from a file)
func indexHandler(w http.ResponseWriter, r *http.Request) {
	index, err := ioutil.ReadFile("../testdata/helm-repo/index.yaml")
	if err != nil {
		panic("Cannot open helm repo test file")
	}
	fmt.Fprint(w, string(index))
}

func brokenIndexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "no yaml here! }")
}

func init() {
	http.HandleFunc("/", http.NotFound)
	http.HandleFunc("/index.yaml", indexHandler)
	http.HandleFunc("/broken-repo/index.yaml", brokenIndexHandler)
	http.HandleFunc("/not-a-repo/index.yaml", http.NotFound)
	go log.Fatal(http.ListenAndServe(":13182", nil))
}

// Negative tests
func TestHelmRepoNotFoundLocal(t *testing.T) {
	h := Helm{
		Repo:  "http://localhost:13182/not-a-repo/",
		Chart: "dependency",
	}

	latestVersion, err := h.LatestVersion()
	require.Error(t, err)
	require.Empty(t, latestVersion)
}

func TestHelmBrokenRepoLocal(t *testing.T) {
	h := Helm{
		Repo:  "http://localhost:13182/broken-repo/",
		Chart: "dependency",
	}

	latestVersion, err := h.LatestVersion()
	require.Error(t, err)
	require.Empty(t, latestVersion)
}

func TestHelmChartNotFoundLocal(t *testing.T) {
	h := Helm{
		Repo:  "http://localhost:13182/",
		Chart: "chart-doesnt-exist",
	}

	latestVersion, err := h.LatestVersion()
	require.Error(t, err)
	require.Empty(t, latestVersion)
}

func TestHelmUnsatisfiableConstraintLocal(t *testing.T) {
	h := Helm{
		Repo:        "http://localhost:13182/",
		Chart:       "dependency",
		Constraints: "> 5.0.0",
	}

	latestVersion, err := h.LatestVersion()
	require.Error(t, err)
	require.Empty(t, latestVersion)
}

// Happy tests
func TestHelmHappyPathLocal(t *testing.T) {
	h := Helm{
		Repo:  "http://localhost:13182/",
		Chart: "dependency",
	}

	latestVersion, err := h.LatestVersion()
	require.NoError(t, err)
	require.NotEmpty(t, latestVersion)
	require.Equal(t, latestVersion, "0.2.0")

	h2 := Helm{
		Repo:  "http://localhost:13182/",
		Chart: "dependency-two",
	}

	latestVersion2, err := h2.LatestVersion()
	require.NoError(t, err)
	require.NotEmpty(t, latestVersion2)
	require.Equal(t, latestVersion2, "2.0.0")
}

func TestHelmHappyPathWithConstraintLocal(t *testing.T) {
	h := Helm{
		Repo:        "http://localhost:13182/",
		Chart:       "dependency",
		Constraints: "< 0.2.0",
	}

	latestVersion, err := h.LatestVersion()
	require.NoError(t, err)
	require.NotEmpty(t, latestVersion)
	require.Equal(t, latestVersion, "0.1.2")
}
