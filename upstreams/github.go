package upstreams

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/blang/semver"
	"github.com/google/go-github/github"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type Github struct {
	UpstreamBase `mapstructure:",squash"`
}

func getClient() *github.Client {
	var client *github.Client
	accessToken := os.Getenv("GITHUB_ACCESS_TOKEN")
	if accessToken != "" {
		log.Debugf("GitHub Access Token provided")
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})
		tc := oauth2.NewClient(oauth2.NoContext, ts)
		client = github.NewClient(tc)
	} else {
		log.Warnf("No GitHub Access Token provided, might run into API limits. Set an access token with the GITHUB_ACCESS_TOKEN env var.")
		client = github.NewClient(nil)
	}
	return client
}

func (upstream Github) LatestVersion() (string, error) {
	log.Debugf("Using GitHub flavour")
	return latestVersion(upstream, getClient)
}

func latestVersion(upstream Github, getClient func() *github.Client) (string, error) {
	client := getClient()
	if !strings.Contains(upstream.URL, "/") {
		return "", fmt.Errorf("Invalid github repo: %v\nGithub repo should be in the form owner/repo, e.g. kubernetes/kubernetes\n", upstream.URL)
	}

	semverConstraints := upstream.Constraints
	if semverConstraints == "" {
		// If no range is passed, just use the broadest possible range
		semverConstraints = ">= 0.0.0"
	}
	expectedRange, err := semver.ParseRange(semverConstraints)
	if err != nil {
		return "", fmt.Errorf("Invalid semver constraints range: %v\n", upstream.Constraints)
	}

	splitUrl := strings.Split(upstream.URL, "/")
	owner := splitUrl[0]
	repo := splitUrl[1]
	opt := &github.ListOptions{Page: 1, PerPage: 20}
	releases, _, err := client.Repositories.ListReleases(context.Background(), owner, repo, opt)

	if err != nil {
		return "", fmt.Errorf("Cannot list releases for repository %v/%v, error: %v\n", owner, repo, err)
	}

	for _, release := range releases {
		if release.TagName == nil {
			log.Debugf("Skipping release without TagName")
		}
		tag := *release.TagName
		if release.Draft != nil && *release.Draft {
			log.Debugf("Skipping draft release: %v\n", tag)
			continue
		}
		if release.Prerelease != nil && *release.Prerelease {
			log.Debugf("Skipping prerelease release: %v\n", tag)
			continue
		}

		// Try to match semver and range
		version, err := semver.Parse(strings.Trim(tag, "v"))
		if err != nil {
			log.Debugf("Version %v is non-semver, cannot validate constraints", tag)
		} else {
			if !expectedRange(version) {
				log.Debugf("Skipping release not matching range constraints (%v): %v\n", upstream.Constraints, tag)
				continue
			}
		}

		log.Debugf("Found latest matching release: %v\n", version)
		return version.String(), nil
	}

	// No latest version found – no versions? Only prereleases?
	return "", fmt.Errorf("No potential version found")
}
