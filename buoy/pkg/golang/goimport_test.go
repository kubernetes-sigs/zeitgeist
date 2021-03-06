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

package golang

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"golang.org/x/net/html"
)

func TestMetaImport_OrgRepo(t *testing.T) {
	tests := map[string]struct {
		meta *MetaImport
		org  string
		repo string
	}{
		"github": {
			meta: &MetaImport{
				RepoRoot: "https://github.com/n3wscott/buoy",
			},
			org:  "n3wscott",
			repo: "buoy",
		},
		"github.git": {
			meta: &MetaImport{
				RepoRoot: "https://github.com/n3wscott/buoy.git",
			},
			org:  "n3wscott",
			repo: "buoy",
		},
		"gitlab": {
			meta: &MetaImport{
				RepoRoot: "http://gitlab.com/repo/oldscott/boiii",
			},
			org:  "oldscott",
			repo: "boiii",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			org, repo := tt.meta.OrgRepo()
			require.Equal(t, org, tt.org)
			require.Equal(t, repo, tt.repo)
		})
	}
}

func TestMetaImport_OrgRepo_UnknownVCS(t *testing.T) {
	meta := &MetaImport{
		RepoRoot: "https://github.com",
	}

	defer func() {
		r := recover()
		// expect panic.
		require.NotNil(t, r)
	}()

	org, repo := meta.OrgRepo()
	// if we get here, it is a fail.
	t.Errorf("Expected OrgRepo to panic, got: %s, %s", org, repo)
}

func TestMetaContent(t *testing.T) {
	tests := map[string]struct {
		meta    string
		doc     *html.Node
		want    string
		wantErr bool
	}{
		"foo meta": {
			meta: "foo",
			doc: func() *html.Node {
				body := `<html><head><meta name="foo" content="bar"></head></html>`
				doc, _ := html.Parse(strings.NewReader(body)) // nolint: errcheck
				return doc
			}(),
			want: "bar",
		},
		"not found": {
			meta: "bar",
			doc: func() *html.Node {
				body := `<html><head><meta name="foo" content="bar"></head></html>`
				doc, _ := html.Parse(strings.NewReader(body)) // nolint: errcheck
				return doc
			}(),
			wantErr: true,
		},
		"go-import": {
			meta: "go-import",
			doc: func() *html.Node {
				body := `<html>
				<head>
					<meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
					<meta name="go-import" content="tableflip.dev/buoy git https://github.com/n3wscott/buoy">
					<meta name="go-source" content="tableflip.dev/buoy https://github.com/n3wscott/buoy https://github.com/n3wscott/buoy/tree/master{/dir} https://github.com/n3wscott/buoy/blob/master{/dir}/{file}#L{line}">
					<meta http-equiv="refresh" content="0; url=https://pkg.go.dev/tableflip.dev/buoy/">
				</head>
				</html>`
				doc, _ := html.Parse(strings.NewReader(body)) // nolint: errcheck
				return doc
			}(),
			want: "tableflip.dev/buoy git https://github.com/n3wscott/buoy",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := metaContent(tt.doc, tt.meta)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, got, tt.want)
		})
	}
}

func TestGetMetaImport(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// nolint: errcheck
		w.Write([]byte(`<html>
		<head>
			<meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
			<meta name="go-import" content="tableflip.dev/buoy git https://github.com/n3wscott/buoy">
			<meta name="go-source" content="tableflip.dev/buoy https://github.com/n3wscott/buoy https://github.com/n3wscott/buoy/tree/master{/dir} https://github.com/n3wscott/buoy/blob/master{/dir}/{file}#L{line}">
			<meta http-equiv="refresh" content="0; url=https://pkg.go.dev/tableflip.dev/buoy/">
		</head>
		</html>`))
	}))
	defer ts.Close()

	meta, err := GetMetaImport(ts.URL)
	require.NoError(t, err)
	require.Equal(t, meta.Prefix, "tableflip.dev/buoy")
	require.Equal(t, meta.VCS, "git")
	require.Equal(t, meta.RepoRoot, "https://github.com/n3wscott/buoy")
}

func TestGetMetaImport_InvalidHost(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	ts.Close()

	_, err := GetMetaImport(ts.URL)
	require.Error(t, err)
}

func TestGetMetaImport_MissingGoImport(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>hi</html>`)) // nolint: errcheck
	}))
	defer ts.Close()

	_, err := GetMetaImport(ts.URL)
	require.Error(t, err)
}
