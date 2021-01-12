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
	"errors"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/net/html"

	"sigs.k8s.io/zeitgeist/buoy/pkg/git"
)

// MetaImport represents the parsed <meta name="go-import"
// content="prefix vcs reporoot" /> tags from HTML files.
type MetaImport struct {
	Prefix, VCS, RepoRoot string
}

func (m *MetaImport) OrgRepo() (org, repo string) {
	repoRoot := strings.TrimSuffix(m.RepoRoot, ".git")
	urlParts := strings.Split(repoRoot, "://")
	parts := strings.Split(urlParts[len(urlParts)-1], "/")

	if len(parts) >= 2 {
		return parts[len(parts)-2], parts[len(parts)-1]
	}

	panic("unknown repo root: " + m.RepoRoot)
}

func metaContent(doc *html.Node, name string) (string, error) {
	var meta *html.Node
	var crawler func(*html.Node)
	crawler = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "meta" {
			for _, attr := range node.Attr {
				if attr.Key == "name" && attr.Val == name {
					meta = node
					return
				}
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			crawler(child)
		}
	}

	crawler(doc)
	if meta != nil {
		for _, attr := range meta.Attr {
			if attr.Key == "content" {
				return attr.Val, nil
			}
		}
	}

	return "", fmt.Errorf("missing <meta name=%s> in the node tree", name)
}

// GetMetaImport fetches and parses header tags named go-import into a
// MetaImport object.
func GetMetaImport(url string) (*MetaImport, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	content, err := metaContent(doc, "go-import")
	if err != nil {
		return nil, err
	}

	f := strings.Fields(content)

	return &MetaImport{
		Prefix:   f[0],
		VCS:      f[1],
		RepoRoot: f[2],
	}, nil
}

// ModuleToRepo resolves a go module name to a remote git repo.
func ModuleToRepo(module string) (*git.Repo, error) {
	url := fmt.Sprintf("https://%s?go-get=1", module)
	meta, err := GetMetaImport(url)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch go import %s: %w", url, err)
	}

	if meta.VCS != "git" {
		return nil, errors.New("unknown VCS: " + meta.VCS)
	}

	return git.GetRepo(module, meta.RepoRoot)
}
