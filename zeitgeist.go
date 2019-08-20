// verify that dependencies are up-to-date across different files
package main

import (
	"bufio"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pluies/zeitgeist/upstreams"

	"github.com/blang/semver"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type Dependencies struct {
	Dependencies []*Dependency `yaml:"dependencies"`
}

type Dependency struct {
	Name     string     `yaml:"name"`
	Version  string     `yaml:"version"`
	Upstream *Upstream  `yaml:"upstream"`
	Semver   bool       `yaml:"semver"`
	RefPaths []*RefPath `yaml:"refPaths"`
}

type RefPath struct {
	Path  string `yaml:"path"`
	Match string `yaml:"match"`
}

type Upstream struct {
	Flavour     upstreams.UpstreamFlavour `yaml:"flavour"`
	URL         string                    `yaml:"url"`
	Constraints string                    `yaml:"constraints"`
}

func main() {
	logLevelStr, ok := os.LookupEnv("LOG_LEVEL")
	if !ok {
		// No LOG_LEVEL env var: default to info
		logLevelStr = "info"
	}
	logLevel, err := log.ParseLevel(logLevelStr)
	if err != nil {
		log.Fatal("Invalid LOG_LEVEL: " + logLevelStr)
	}
	log.SetLevel(logLevel)

	flag.Parse()
	args := flag.Args()
	if len(args) != 1 {
		log.Fatalf("usage: dependency <file>")
	}

	depFileStr := args[0]
	depFile, err := ioutil.ReadFile(depFileStr)
	if err != nil {
		log.Fatal(err)
	}

	base := filepath.Dir(depFileStr)
	mismatchErrorMessage := "ERROR: %v indicates that %v should be at version %v, but the following files didn't match:\n\n" +
		"%v\n\nif you are changing the version of %v, make sure all of the following files are updated with the newest version including %v\n" +
		"then run ./hack/verify-external-dependencies-version.sh\n\n"
	externalDeps := &Dependencies{}
	var pathsToUpdate []string
	err = yaml.Unmarshal(depFile, externalDeps)
	if err != nil {
		log.Fatal(err)
	}

	for _, dep := range externalDeps.Dependencies {
		log.Debugf("Examining dependency: %v", dep.Name)
		for _, refPath := range dep.RefPaths {
			filePath := filepath.Join(base, refPath.Path)
			file, err := os.Open(filePath)
			if err != nil {
				log.Fatalf("Error opening %v: %v", filePath, err)
			}
			log.Debugf("Examining file: %v", filePath)
			match := refPath.Match
			matcher := regexp.MustCompile(match)
			scanner := bufio.NewScanner(file)

			var found bool
			var lineNumber int
			for scanner.Scan() {
				lineNumber += 1
				line := scanner.Text()
				if matcher.MatchString(line) {
					if strings.Contains(line, dep.Version) {
						log.Debugf("Line %v matches expected regexp (%v) and version (%v): %v", lineNumber, match, dep.Version, line)
						found = true
						break
					} else {
						log.Warnf("Line %v matches expected regexp (%v), but not version (%v): %v", lineNumber, match, dep.Version, line)
					}
				}
			}
			if !found {
				log.Debugf("Finished reading file %v, no match found.", filePath)
				pathsToUpdate = append(pathsToUpdate, refPath.Path)
			}
		}
		if len(pathsToUpdate) > 0 {
			log.Fatalf(mismatchErrorMessage, depFileStr, dep.Name, dep.Version, strings.Join(pathsToUpdate, "\n"), dep.Name, depFileStr)
		}

		if dep.Upstream == nil {
			continue
		}

		var latestVersionString string = dep.Version
		var versionString string = dep.Version

		if dep.Upstream.Flavour == upstreams.GitHub {
			gh := upstreams.Github{
				URL:         dep.Upstream.URL,
				Constraints: dep.Upstream.Constraints,
			}
			latestVersionString = gh.LatestVersion()
		}

		latestV := Version(latestVersionString)
		versionV := Version(versionString)
		if latestV.MoreRecentThan(versionV) {
			log.Infof("Update available for dependency %v: %v (current: %v)\n", dep.Name, latestV, versionV)
		} else {
			log.Infof("No update available for dependency %v: %v (latest: %v)\n", dep.Name, versionV, latestV)
		}
	}
}

type Version string

func (a Version) MoreRecentThan(b Version) bool {
	// Try and parse as Semver first
	semverComparison := true
	aSemver, err := semver.Parse(string(a))
	if err != nil {
		semverComparison = false
	}
	bSemver, err := semver.Parse(string(b))
	if err != nil {
		semverComparison = false
	}
	if semverComparison {
		return aSemver.GT(bSemver)
	} else {
		// Failed semver: fallback to standard string comparison (lexicographic)
		return strings.Compare(string(a), string(b)) > 0
	}
}
