package upstreams

import (
	"context"
	"strings"

	"github.com/blang/semver"
	"github.com/google/go-github/github"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type Github struct {
	AccessToken string
	URL         string
	Constraints string
}

func (upstream Github) LatestVersion() string {
	log.Debugf("Using GitHub flavour")

	if !strings.Contains(upstream.URL, "/") {
		log.Fatalf("Invalid github repo: %v\nGithub repo should be in the form owner/repo, e.g. kubernetes/kubernetes\n", upstream.URL)
	}

	semverConstraints := upstream.Constraints
	if semverConstraints == "" {
		// If no range is passed, just use the broadest possible range
		semverConstraints = ">= 0.0.0"
	}
	expectedRange, err := semver.ParseRange(semverConstraints)
	if err != nil {
		log.Fatalf("Invalid semver constraints range: %v\n", upstream.Constraints)
	}

	var client *github.Client
	if upstream.AccessToken != "" {
		log.Debugf("GitHub Access Token provided")
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: upstream.AccessToken})
		tc := oauth2.NewClient(oauth2.NoContext, ts)
		client = github.NewClient(tc)
	} else {
		log.Warnf("No GitHub Access Token provided, might run into API limits. Set an access token with the GITHUB_ACCESS_TOKEN env var.")
		client = github.NewClient(nil)
	}

	splitUrl := strings.Split(upstream.URL, "/")
	owner := splitUrl[0]
	repo := splitUrl[1]
	opt := &github.ListOptions{Page: 1, PerPage: 20}
	releases, _, err := client.Repositories.ListReleases(context.Background(), owner, repo, opt)

	if err != nil {
		log.Fatalf("Cannot list releases for repository %v/%v, error: %v\n", owner, repo, err)
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
		return version.String()
	}

	// No latest version found – no versions? Only prereleases?
	// TODO Handle this case better
	return ""
}
