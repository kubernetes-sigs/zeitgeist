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

// Client holds any client that is needed.
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
	Upgrade(dependencyFilePath, basePath string) ([]string, error)

	SetVersion(dependencyFilePath, basePath, dependency, version string) error

	RemoteExport(dependencyFilePath string) ([]VersionUpdate, error)

	CheckUpstreamVersions(deps []*Dependency) ([]VersionUpdateInfo, error)
}

type UnsupportedError struct {
	message string
}

func (u UnsupportedError) Error() string {
	return u.message
}

// Dependencies is used to deserialise the configuration file.
type Dependencies struct {
	Dependencies []*Dependency `yaml:"dependencies"`
}

// Dependency is the internal representation of a dependency.
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

// RefPath represents a file to check for a reference to the version.
type RefPath struct {
	// Path of the file to test
	Path string `yaml:"path"`
	// Match expression for the line that should contain the dependency's version. Regexp is supported.
	Match string `yaml:"match"`
}

// UnmarshalYAML implements custom unmarshalling of Dependency with validation.
func (decoded *Dependency) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Use a different type to prevent infinite loop in unmarshalling
	type DependencyYAML Dependency

	d := (*DependencyYAML)(decoded)

	if err := unmarshal(&d); err != nil {
		return err
	}

	// Custom validation for the Dependency type
	if d.Name == "" {
		return fmt.Errorf("dependency has no `name`: %#v", d)
	}

	if d.Version == "" {
		return fmt.Errorf("dependency has no `version`: %#v", d)
	}

	// Default scheme to Semver if unset
	if d.Scheme == "" {
		d.Scheme = Semver
	}

	// Validate Scheme
	switch d.Scheme {
	case Semver, Alpha, Random:
		// All good!
	default:
		return fmt.Errorf("unknown version scheme: %s", d.Scheme)
	}

	// Validate RefPaths
	for _, refPath := range d.RefPaths {
		if refPath.Path == "" {
			return fmt.Errorf("dependency %s is invalid: refPath is missing `path`", d.Name)
		}
		if refPath.Match == "" {
			return fmt.Errorf("dependency %s is invalid: refPath is missing `match`", d.Name)
		}
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

	decoder := yaml.NewDecoder(bytes.NewReader(depFile))
	decoder.KnownFields(true) // Disallow unknown fields
	if err = decoder.Decode(dependencies); err != nil {
		if err.Error() == "EOF" {
			return nil, fmt.Errorf("can't decode YAML from configuration file %s: %w", dependencyFilePath, err)
		}
		re := regexp.MustCompile(`field (.*) not found`)
		matches := re.FindStringSubmatch(err.Error())
		if len(matches) > 1 {
			return nil, fmt.Errorf("unexpected key: %s", matches[1])
		}
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

// NewClient returns all clients that can be used to the validation.
func NewLocalClient() (Client, error) {
	return &LocalClient{}, nil
}

// LocalCheck checks whether dependencies are in-sync locally
//
// Will return an error if the dependency cannot be found in the files it has defined, or if the version does not match.
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

// SetVersion sets the version of a dependency to the specified version
//
// Will return an error  if updating files fails.
func (c *LocalClient) SetVersion(dependencyFilePath, basePath, dependency, version string) error {
	externalDeps, err := FromFile(dependencyFilePath)
	if err != nil {
		return err
	}

	found := false
	for _, dep := range externalDeps.Dependencies {
		if dep.Name == dependency {
			found = true

			if err := upgradeDependency(basePath, dep, &VersionUpdateInfo{
				Name: dep.Name,
				Current: Version{
					Version: dep.Version,
					Scheme:  dep.Scheme,
				},
				Latest: Version{
					Version: version,
					Scheme:  dep.Scheme,
				},
				UpdateAvailable: true,
			}); err != nil {
				return err
			}

			dep.Version = version
		}
	}

	if !found {
		return fmt.Errorf("dependency %s not found", dependency)
	}

	// Update the dependencies file to reflect the upgrades
	err = ToFile(dependencyFilePath, externalDeps)
	if err != nil {
		return err
	}

	return nil
}

func (c *LocalClient) RemoteCheck(dependencyFilePath string) ([]string, error) { //nolint: revive
	return nil, UnsupportedError{"remote checks are not supported by the local client"}
}

func (c *LocalClient) Upgrade(dependencyFilePath, basePath string) ([]string, error) { //nolint: revive
	return nil, UnsupportedError{"upgrade is not supported by the local client"}
}

func (c *LocalClient) RemoteExport(dependencyFilePath string) ([]VersionUpdate, error) { //nolint: revive
	return nil, UnsupportedError{"remote export is not supported by the local client"}
}

func (c *LocalClient) CheckUpstreamVersions(deps []*Dependency) ([]VersionUpdateInfo, error) { //nolint: revive
	return nil, UnsupportedError{"CheckUpstreamVersions is not supported by the local client"}
}

var NewRemoteClient = func() (Client, error) {
	return nil, UnsupportedError{"remote upstream functionality is not supported by this command; use sigs.k8s.io/zeitgeist/remote/zeitgeist"}
}

func upgradeDependency(basePath string, dependency *Dependency, versionUpdate *VersionUpdateInfo) error {
	log.Debugf("running upgradeDependency, versionUpdate %#v", versionUpdate)
	for _, refPath := range dependency.RefPaths {
		err := replaceInFile(basePath, refPath, versionUpdate)
		if err != nil {
			return err
		}
	}

	return nil
}

func replaceInFile(basePath string, refPath *RefPath, versionUpdate *VersionUpdateInfo) error {
	filename := filepath.Join(basePath, refPath.Path)
	log.Debugf("running replaceInFile, refpath is %#v, versionUpdate %#v", refPath, versionUpdate)

	matcher, err := regexp.Compile(refPath.Match)
	if err != nil {
		return fmt.Errorf("compiling regex: %w", err)
	}

	inputFile, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	lines := strings.Split(string(inputFile), "\n")

	for i, line := range lines {
		if matcher.MatchString(line) {
			if strings.Contains(line, versionUpdate.Current.Version) {
				log.Debugf(
					"Line %d matches expected regexp %q and version %q: %s",
					i,
					refPath.Match,
					versionUpdate.Current.Version,
					line,
				)

				// The actual upgrade:
				lines[i] = strings.ReplaceAll(line, versionUpdate.Current.Version, versionUpdate.Latest.Version)
			}
		}
	}

	upgradedFile := strings.Join(lines, "\n")

	// Finally, write the file out
	err = os.WriteFile(filename, []byte(upgradedFile), 0o644)
	if err != nil {
		return fmt.Errorf("writing file: %w", err)
	}
	return nil
}
