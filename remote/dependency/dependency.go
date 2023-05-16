/*
Copyright 2023 The Kubernetes Authors.

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
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"

	deppkg "sigs.k8s.io/zeitgeist/dependency"
	"sigs.k8s.io/zeitgeist/upstream"
)

func init() {
	deppkg.NewRemoteClient = NewRemoteClient
}

type RemoteClient struct {
	LocalClient  deppkg.Client
	AWSEC2Client ec2iface.EC2API
}

type versionUpdateInfo struct {
	name            string
	current         deppkg.Version
	latest          deppkg.Version
	updateAvailable bool
}

func NewRemoteClient() (deppkg.Client, error) {
	localClient, err := deppkg.NewLocalClient()
	if err != nil {
		return nil, err
	}
	return &RemoteClient{
		LocalClient:  localClient,
		AWSEC2Client: upstream.NewAWSClient(),
	}, nil
}

func (c *RemoteClient) LocalCheck(dependencyFilePath, basePath string) error {
	return c.LocalClient.LocalCheck(dependencyFilePath, basePath)
}

// RemoteCheck checks whether dependencies are up to date with upstream
//
// Will return an error if checking the versions upstream fails.
//
// Out-of-date dependencies will be printed out on stdout at the INFO level.
func (c *RemoteClient) RemoteCheck(dependencyFilePath string) ([]string, error) {
	externalDeps, err := deppkg.FromFile(dependencyFilePath)
	if err != nil {
		return nil, err
	}

	updates := make([]string, 0)

	versionUpdateInfos, err := c.checkUpstreamVersions(externalDeps.Dependencies)
	if err != nil {
		return nil, err
	}

	for _, vu := range versionUpdateInfos {
		if vu.updateAvailable {
			updates = append(
				updates,
				fmt.Sprintf(
					"Update available for dependency %s: %s (current: %s)",
					vu.name,
					vu.latest.Version,
					vu.current.Version,
				),
			)
		} else {
			log.Debugf(
				"No update available for dependency %s: %s (latest: %s)\n",
				vu.name,
				vu.current.Version,
				vu.latest.Version,
			)
		}
	}

	return updates, nil
}

// Upgrade retrieves the most up-to-date version of the dependency and replaces
// the local version with the most up-to-date version.
//
// Will return an error if checking the versions upstream fails, or if updating
// files fails.
func (c *RemoteClient) Upgrade(dependencyFilePath string) ([]string, error) {
	externalDeps, err := deppkg.FromFile(dependencyFilePath)
	if err != nil {
		return nil, err
	}

	upgrades := make([]string, 0)
	upgradedDependencies := make([]*deppkg.Dependency, 0)

	versionUpdateInfos, err := c.checkUpstreamVersions(externalDeps.Dependencies)
	if err != nil {
		return nil, err
	}

	for _, vu := range versionUpdateInfos {
		dependency, err := findDependencyByName(externalDeps.Dependencies, vu.name)
		if err != nil {
			return nil, err
		}

		if vu.updateAvailable {
			err = upgradeDependency(dependency, &vu)
			if err != nil {
				return nil, err
			}

			dependency.Version = vu.latest.Version
			upgradedDependencies = append(
				upgradedDependencies,
				dependency,
			)

			upgrades = append(
				upgrades,
				fmt.Sprintf(
					"Upgraded dependency %s from version %s to version %s",
					vu.name,
					vu.current.Version,
					vu.latest.Version,
				),
			)
		} else {
			upgradedDependencies = append(
				upgradedDependencies,
				dependency,
			)

			log.Debugf(
				"No update available for dependency %s: %s (latest: %s)\n",
				vu.name,
				vu.current.Version,
				vu.latest.Version,
			)
		}
	}

	// Update the dependencies file to reflect the upgrades
	err = deppkg.ToFile(dependencyFilePath, &deppkg.Dependencies{
		Dependencies: upgradedDependencies,
	})
	if err != nil {
		return nil, err
	}

	return upgrades, nil
}

func findDependencyByName(dependencies []*deppkg.Dependency, name string) (*deppkg.Dependency, error) {
	for _, dep := range dependencies {
		if dep.Name == name {
			return dep, nil
		}
	}
	return nil, fmt.Errorf("cannot find dependency by name: %s", name)
}

func upgradeDependency(dependency *deppkg.Dependency, versionUpdate *versionUpdateInfo) error {
	log.Debugf("running upgradeDependency, versionUpdate %#v", versionUpdate)
	for _, refPath := range dependency.RefPaths {
		err := replaceInFile(refPath, versionUpdate)
		if err != nil {
			return err
		}
	}

	return nil
}

func replaceInFile(refPath *deppkg.RefPath, versionUpdate *versionUpdateInfo) error {
	log.Debugf("running replaceInFile, refpath is %#v, versionUpdate %#v", refPath, versionUpdate)

	matcher, err := regexp.Compile(refPath.Match)
	if err != nil {
		return fmt.Errorf("compiling regex: %w", err)
	}

	inputFile, err := os.ReadFile(refPath.Path)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	lines := strings.Split(string(inputFile), "\n")

	for i, line := range lines {
		if matcher.MatchString(line) {
			if strings.Contains(line, versionUpdate.current.Version) {
				log.Debugf(
					"Line %d matches expected regexp %q and version %q: %s",
					i,
					refPath.Match,
					versionUpdate.current.Version,
					line,
				)

				// The actual upgrade:
				lines[i] = strings.ReplaceAll(line, versionUpdate.current.Version, versionUpdate.latest.Version)
			}
		}
	}

	upgradedFile := strings.Join(lines, "\n")

	// Finally, write the file out
	err = os.WriteFile(refPath.Path, []byte(upgradedFile), 0o644)

	if err != nil {
		return fmt.Errorf("writing file: %w", err)
	}
	return nil
}

func (c *RemoteClient) RemoteExport(dependencyFilePath string) ([]deppkg.VersionUpdate, error) {
	externalDeps, err := deppkg.FromFile(dependencyFilePath)
	if err != nil {
		return nil, err
	}

	versionUpdates := []deppkg.VersionUpdate{}

	versionUpdatesInfos, err := c.checkUpstreamVersions(externalDeps.Dependencies)
	if err != nil {
		return nil, err
	}

	for _, vui := range versionUpdatesInfos {
		if vui.updateAvailable {
			versionUpdates = append(versionUpdates, deppkg.VersionUpdate{
				Name:       vui.name,
				Version:    vui.current.Version,
				NewVersion: vui.latest.Version,
			})
		} else {
			log.Debugf(
				"No update available for dependency %s: %s (latest: %s)\n",
				vui.name,
				vui.current.Version,
				vui.latest.Version,
			)
		}
	}
	return versionUpdates, nil
}

func (c *RemoteClient) checkUpstreamVersions(deps []*deppkg.Dependency) ([]versionUpdateInfo, error) {
	versionUpdates := []versionUpdateInfo{}
	for _, dep := range deps {
		if dep.Upstream == nil {
			continue
		}

		up := dep.Upstream
		latestVersion := deppkg.Version{Version: dep.Version, Scheme: dep.Scheme}
		currentVersion := deppkg.Version{Version: dep.Version, Scheme: dep.Scheme}

		var err error

		// Cast the flavour from the currently unknown upstream type
		flavour := upstream.Flavour(up["flavour"])
		switch flavour {
		case upstream.DummyFlavour:
			var d upstream.Dummy

			decodeErr := mapstructure.Decode(up, &d)
			if decodeErr != nil {
				return nil, decodeErr
			}

			latestVersion.Version, err = d.LatestVersion()
		case upstream.GithubFlavour:
			var gh upstream.Github

			decodeErr := mapstructure.Decode(up, &gh)
			if decodeErr != nil {
				return nil, decodeErr
			}

			latestVersion.Version, err = gh.LatestVersion()
		case upstream.GitLabFlavour:
			var gl upstream.GitLab

			decodeErr := mapstructure.Decode(up, &gl)
			if decodeErr != nil {
				return nil, decodeErr
			}

			latestVersion.Version, err = gl.LatestVersion()
		case upstream.HelmFlavour:
			var h upstream.Helm

			decodeErr := mapstructure.Decode(up, &h)
			if decodeErr != nil {
				return nil, decodeErr
			}

			latestVersion.Version, err = h.LatestVersion()
		case upstream.AMIFlavour:
			var ami upstream.AMI

			decodeErr := mapstructure.Decode(up, &ami)
			if decodeErr != nil {
				return nil, decodeErr
			}

			ami.ServiceClient = c.AWSEC2Client

			latestVersion.Version, err = ami.LatestVersion()
		case upstream.ContainerFlavour:
			var ct upstream.Container

			decodeErr := mapstructure.Decode(up, &ct)
			if decodeErr != nil {
				log.Debug("errr decoding")
				return nil, decodeErr
			}

			latestVersion.Version, err = ct.LatestVersion()
		case upstream.EKSFlavour:
			var eks upstream.EKS

			decodeErr := mapstructure.Decode(up, &eks)
			if decodeErr != nil {
				return nil, decodeErr
			}

			latestVersion.Version, err = eks.LatestVersion()
		default:
			return nil, fmt.Errorf("unknown upstream flavour '%#v' for dependency %s", flavour, dep.Name)
		}

		if err != nil {
			return nil, err
		}

		updateAvailable, err := latestVersion.MoreSensitivelyRecentThan(currentVersion, dep.Sensitivity)
		if err != nil {
			return nil, err
		}

		versionUpdates = append(versionUpdates, versionUpdateInfo{
			name:            dep.Name,
			current:         currentVersion,
			latest:          latestVersion,
			updateAvailable: updateAvailable,
		})
	}

	return versionUpdates, nil
}
