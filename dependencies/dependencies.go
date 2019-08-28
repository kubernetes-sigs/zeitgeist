package dependencies

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pluies/zeitgeist/upstreams"

	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type Dependencies struct {
	Dependencies []*Dependency `yaml:"dependencies"`
}

type Dependency struct {
	Name string `yaml:"name"`
	// Version of the dependency that should be present throughout your code
	Version string `yaml:"version"`
	// Scheme for versioning.
	// Supported values:
	// - `semver`: [Semantic versioning](https://semver.org/), default
	// - `alpha`: alphanumeric, will use standard string sorting
	// - `random`: e.g. when releases are hashes and do not support sorting
	Scheme   VersionScheme     `yaml:"scheme"`
	Upstream map[string]string `yaml:"upstream"`
	// List of references to this dependency in local files
	RefPaths []*RefPath `yaml:"refPaths"`
}

type RefPath struct {
	// Path of the file to test
	Path string `yaml:"path"`
	// Match expression for the line that should contain the dependency's version. Regexp is supported.
	Match string `yaml:"match"`
}

// Custom unmarshalling of Dependency to add extra validation
func (decoded *Dependency) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Use a different type to prevent infinite loop in unmarshalling
	type DependencyYAML Dependency
	d := (*DependencyYAML)(decoded)
	if err := unmarshal(&d); err != nil {
		return err
	}
	// Custom validation for the Dependency type
	if d.Name == "" {
		return fmt.Errorf("Dependency has no `name`: %v", d)
	}
	if d.Version == "" {
		return fmt.Errorf("Dependency has no `version`: %v", d)
	}
	// Default scheme to Semver if unset
	if d.Scheme == "" {
		d.Scheme = Semver
	}
	// Validate Scheme and return
	switch d.Scheme {
	case Semver, Alpha, Random:
		// All good!
	default:
		return fmt.Errorf("Unknown version scheme: %s", d.Scheme)
	}
	log.Debugf("Deserialised Dependency %v: %v", d.Name, d)
	return nil
}

func fromFile(dependencyFilePath string) (*Dependencies, error) {
	depFile, err := ioutil.ReadFile(dependencyFilePath)
	if err != nil {
		return nil, err
	}

	dependencies := &Dependencies{}
	err = yaml.Unmarshal(depFile, dependencies)
	if err != nil {
		return nil, err
	}
	return dependencies, nil
}

func LocalCheck(dependencyFilePath string) error {
	base := filepath.Dir(dependencyFilePath)
	externalDeps, err := fromFile(dependencyFilePath)
	if err != nil {
		return err
	}
	var nonMatchingPaths []string
	for _, dep := range externalDeps.Dependencies {
		log.Debugf("Examining dependency: %v", dep.Name)
		for _, refPath := range dep.RefPaths {
			filePath := filepath.Join(base, refPath.Path)
			file, err := os.Open(filePath)
			if err != nil {
				log.Errorf("Error opening %v: %v", filePath, err)
				return err
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
						log.Debugf("Line %v matches expected regexp '%v' and version '%v':\n%v", lineNumber, match, dep.Version, line)
						found = true
						break
					} else {
						log.Warnf("Line %v matches expected regexp '%v', but not version '%v':\n%v", lineNumber, match, dep.Version, line)
					}
				}
			}
			if !found {
				log.Debugf("Finished reading file %v, no match found.", filePath)
				nonMatchingPaths = append(nonMatchingPaths, refPath.Path)
			}
		}

		if len(nonMatchingPaths) > 0 {
			log.Errorf("%v indicates that %v should be at version %v, but the following files didn't match:\n"+
				"%v\n", dependencyFilePath, dep.Name, dep.Version, strings.Join(nonMatchingPaths, "\n"))
			return errors.New("Dependencies are not in sync")
		}
	}
	return nil
}

func RemoteCheck(dependencyFilePath string) error {
	externalDeps, err := fromFile(dependencyFilePath)
	if err != nil {
		return err
	}
	for _, dep := range externalDeps.Dependencies {
		log.Debugf("Examining dependency: %v", dep.Name)

		if dep.Upstream == nil {
			continue
		}
		upstream := dep.Upstream

		latestVersion := Version{dep.Version, dep.Scheme}
		currentVersion := Version{dep.Version, dep.Scheme}
		var err error

		// Cast the flavour from the currently unknown upstream type
		flavour := upstreams.UpstreamFlavour(upstream["flavour"])
		switch flavour {
		case upstreams.DummyFlavour:
			var d upstreams.Dummy
			decodeErr := mapstructure.Decode(upstream, &d)
			if decodeErr != nil {
				return decodeErr
			}
			latestVersion.Version, err = d.LatestVersion()
		case upstreams.GithubFlavour:
			var gh upstreams.Github
			decodeErr := mapstructure.Decode(upstream, &gh)
			if decodeErr != nil {
				return decodeErr
			}
			latestVersion.Version, err = gh.LatestVersion()
		case upstreams.AMIFlavour:
			var ami upstreams.AMI
			decodeErr := mapstructure.Decode(upstream, &ami)
			if decodeErr != nil {
				return decodeErr
			}
			latestVersion.Version, err = ami.LatestVersion()
		default:
			return fmt.Errorf("Unknown upstream flavour '%v' for dependency %v", flavour, dep.Name)
		}
		if err != nil {
			return err
		}

		updateAvailable, err := latestVersion.MoreRecentThan(currentVersion)
		if err != nil {
			return err
		}

		if updateAvailable {
			log.Infof("Update available for dependency %v: %v (current: %v)\n", dep.Name, latestVersion.Version, currentVersion.Version)
		} else {
			log.Infof("No update available for dependency %v: %v (latest: %v)\n", dep.Name, currentVersion.Version, latestVersion.Version)
		}
	}
	return nil
}
