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

// Package dependencies checks dependencies, locally or remotely
package dependency

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// Client holds any client that is needed
type Client interface {
	// LocalCheck checks whether dependencies are in-sync locally
	//
	// Will return an error if the dependency cannot be found in the files it has defined, or if the version does not match
	LocalCheck(dependencyFilePath, basePath string) error

	// RemoteCheck checks whether dependencies are up to date with upstream
	//
	// Will return an error if checking the versions upstream fails.
	//
	// Out-of-date dependencies will be printed out on stdout at the INFO level.
	RemoteCheck(dependencyFilePath string) ([]string, error)

	// Upgrade retrieves the most up-to-date version of the dependency and replaces
	// the local version with the most up-to-date version.
	//
	// Will return an error if checking the versions upstream fails, or if updating
	// files fails.
	Upgrade(dependencyFilePath string) ([]string, error)

	RemoteExport(dependencyFilePath string) ([]VersionUpdate, error)
}

type UnsupportedError struct {
	message string
}

func (u UnsupportedError) Error() string {
	return u.message
}

// Dependencies is used to deserialise the configuration file
type Dependencies struct {
	Dependencies []*Dependency `yaml:"dependencies"`
}

// Dependency is the internal representation of a dependency
type Dependency struct {
	Name string `yaml:"name"`
	// Version of the dependency that should be present throughout your code
	Version string `yaml:"version"`
	// Scheme for versioning this dependency
	Scheme VersionScheme `yaml:"scheme"`
	// Optional: sensitivity, to alert e.g. on new major versions
	Sensitivity VersionSensitivity `yaml:"sensitivity,omitempty"`
	// Optional: upstream
	Upstream map[string]string `yaml:"upstream,omitempty"`
	// List of references to this dependency in local files
	RefPaths []*RefPath `yaml:"refPaths"`
}

// RefPath represents a file to check for a reference to the version
type RefPath struct {
	// Path of the file to test
	Path string `yaml:"path"`
	// Match expression for the line that should contain the dependency's version. Regexp is supported.
	Match string `yaml:"match"`
}

// UnmarshalYAML implements custom unmarshalling of Dependency with validation
func (decoded *Dependency) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Use a different type to prevent infinite loop in unmarshalling
	type DependencyYAML Dependency

	d := (*DependencyYAML)(decoded)

	if err := unmarshal(&d); err != nil {
		return err
	}

	// Custom validation for the Dependency type
	if d.Name == "" {
		return fmt.Errorf("Dependency has no `name`: %#v", d)
	}

	if d.Version == "" {
		return fmt.Errorf("Dependency has no `version`: %#v", d)
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
		return fmt.Errorf("unknown version scheme: %s", d.Scheme)
	}

	log.Debugf("Deserialised Dependency %s: %#v", d.Name, d)

	return nil
}

func FromFile(dependencyFilePath string) (*Dependencies, error) {
	depFile, err := os.ReadFile(dependencyFilePath)
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

func ToFile(dependencyFilePath string, dependencies *Dependencies) error {
	var output bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&output)
	yamlEncoder.SetIndent(2)

	err := yamlEncoder.Encode(dependencies)
	if err != nil {
		return err
	}

	err = os.WriteFile(dependencyFilePath, output.Bytes(), 0o644)
	if err != nil {
		return err
	}

	return nil
}

type LocalClient struct{}

// NewClient returns all clients that can be used to the validation
func NewLocalClient() (Client, error) {
	return &LocalClient{}, nil
}

// LocalCheck checks whether dependencies are in-sync locally
//
// Will return an error if the dependency cannot be found in the files it has defined, or if the version does not match
func (c *LocalClient) LocalCheck(dependencyFilePath, basePath string) error {
	log.Debugf("Base path: %s", basePath)
	externalDeps, err := FromFile(dependencyFilePath)
	if err != nil {
		return err
	}

	var nonMatchingPaths []string
	for _, dep := range externalDeps.Dependencies {
		log.Debugf("Examining dependency: %s", dep.Name)

		for _, refPath := range dep.RefPaths {
			filePath := filepath.Join(basePath, refPath.Path)

			log.Debugf("Examining file: %s", filePath)

			file, err := os.Open(filePath)
			if err != nil {
				return err
			}

			match := refPath.Match
			matcher, err := regexp.Compile(match)
			if err != nil {
				return fmt.Errorf("compiling regex: %w", err)
			}
			scanner := bufio.NewScanner(file)

			var found bool

			var lineNumber int
			for scanner.Scan() {
				lineNumber++

				line := scanner.Text()
				if matcher.MatchString(line) {
					if strings.Contains(line, dep.Version) {
						log.Debugf(
							"Line %d matches expected regexp %q and version %q: %s",
							lineNumber,
							match,
							dep.Version,
							line,
						)

						found = true
						break
					}
				}
			}

			if !found {
				log.Debugf("Finished reading file %s, no match found.", filePath)

				nonMatchingPaths = append(nonMatchingPaths, refPath.Path)
			}
		}

		if len(nonMatchingPaths) > 0 {
			log.Errorf(
				"%s indicates that %s should be at version %s, but the following files didn't match: %s",
				dependencyFilePath,
				dep.Name,
				dep.Version,
				strings.Join(nonMatchingPaths, ", "),
			)

			return errors.New("Dependencies are not in sync")
		}
	}

	return nil
}

func (c *LocalClient) RemoteCheck(dependencyFilePath string) ([]string, error) {
	return nil, UnsupportedError{"remote checks are not supported by the local client"}
}

func (c *LocalClient) Upgrade(dependencyFilePath string) ([]string, error) {
	return nil, UnsupportedError{"upgrade is not supported by the local client"}
}

func (c *LocalClient) RemoteExport(dependencyFilePath string) ([]VersionUpdate, error) {
	return nil, UnsupportedError{"remote export is not supported by the local client"}
}

var NewRemoteClient = func() (Client, error) {
	return nil, UnsupportedError{"remote upstream functionality is not supported by this command; use sigs.k8s.io/zeitgeist/remote/zeitgeist"}
}
