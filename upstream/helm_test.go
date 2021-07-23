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
	"net/http"
	"net/http/httptest"
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

// We need to set up a local webserver to serve the Helm repo index
// (As far as I can tell, it can't read directly from a file)
// We do that by instantiating an httptest server in each test that requires it
func helmHandler(rw http.ResponseWriter, req *http.Request) {
	url := req.URL.String()
	switch url {
	case "/":
		fmt.Fprint(rw, "zeitgeist testing server")
	case "/index.yaml":
		index, err := ioutil.ReadFile("../testdata/helm-repo/index.yaml")
		if err != nil {
			panic("Cannot open helm repo test file")
		}
		fmt.Fprint(rw, string(index))
	case "/broken-repo/index.yaml":
		fmt.Fprint(rw, "bad yaml here } !")
	default:
		rw.WriteHeader(404)
	}
}

// Negative tests
func TestHelmRepoNotFoundLocal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(helmHandler))
	defer server.Close()

	h := Helm{
		Repo:  server.URL + "/not-a-repo/",
		Chart: "dependency",
	}

	latestVersion, err := h.LatestVersion()
	require.Error(t, err)
	require.Empty(t, latestVersion)
}

func TestHelmBrokenRepoLocal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(helmHandler))
	defer server.Close()

	h := Helm{
		Repo:  server.URL + "/broken-repo/",
		Chart: "dependency",
	}

	latestVersion, err := h.LatestVersion()
	require.Error(t, err)
	require.Empty(t, latestVersion)
}

func TestHelmChartNotFoundLocal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(helmHandler))
	defer server.Close()

	h := Helm{
		Repo:  server.URL,
		Chart: "chart-doesnt-exist",
	}

	latestVersion, err := h.LatestVersion()
	require.Error(t, err)
	require.Empty(t, latestVersion)
}

func TestHelmUnsatisfiableConstraintLocal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(helmHandler))
	defer server.Close()

	h := Helm{
		Repo:        server.URL,
		Chart:       "dependency",
		Constraints: "> 5.0.0",
	}

	latestVersion, err := h.LatestVersion()
	require.Error(t, err)
	require.Empty(t, latestVersion)
}

// Happy tests
func TestHelmHappyPathLocal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(helmHandler))
	defer server.Close()

	h := Helm{
		Repo:  server.URL,
		Chart: "dependency",
	}

	latestVersion, err := h.LatestVersion()
	require.NoError(t, err)
	require.NotEmpty(t, latestVersion)
	require.Equal(t, latestVersion, "0.2.0")

	h2 := Helm{
		Repo:  server.URL,
		Chart: "dependency-two",
	}

	latestVersion2, err := h2.LatestVersion()
	require.NoError(t, err)
	require.NotEmpty(t, latestVersion2)
	require.Equal(t, latestVersion2, "2.0.0")
}

func TestHelmHappyPathWithConstraintLocal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(helmHandler))
	defer server.Close()

	h := Helm{
		Repo:        server.URL,
		Chart:       "dependency",
		Constraints: "< 0.2.0",
	}

	latestVersion, err := h.LatestVersion()
	require.NoError(t, err)
	require.NotEmpty(t, latestVersion)
	require.Equal(t, latestVersion, "0.1.2")
}
