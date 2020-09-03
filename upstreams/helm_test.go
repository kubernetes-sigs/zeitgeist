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

package upstreams

import (
	"testing"
)

func TestHelmHappyPath(t *testing.T) {
	helm := Helm{
		Name: "fluentd",
	}
	helm1 := Helm{
		Name:        "fluentd",
		Constraints: "< 2.0.0",
	}
	latestVersion, err := helm.LatestVersion()
	if err != nil {
		t.Errorf("Failed Helm happy path test: %v", err)
	}
	if latestVersion == "" {
		t.Errorf("Got an empty latestVersion")
	}

	latestVersion1, err1 := helm1.LatestVersion()
	if err1 != nil {
		t.Errorf("Failed Helm happy path test: %v", err1)
	}
	if latestVersion1 == "" {
		t.Errorf("Got an empty latestVersion")
	}

	if latestVersion == latestVersion1 {
		t.Errorf("Got the same latestVersion with or without constraints")
	}
}

func TestHelmBrokenRepo(t *testing.T) {
	helm := Helm{
		Repo: "https://example.com/",
		Name: "fluentd",
	}
	_, err := helm.LatestVersion()
	if err == nil {
		t.Errorf("Should have failed on broken repo")
	}
}

func TestHelmBadChart(t *testing.T) {
	helm := Helm{
		Repo: "stable",
		Name: "this-chart-does-not-exist",
	}
	_, err := helm.LatestVersion()
	if err == nil {
		t.Errorf("Should have failed on broken chart")
	}
}

func TestHelmBadConstraint(t *testing.T) {
	helm := Helm{
		Repo:        "stable",
		Name:        "fluentd",
		Constraints: ">2500.0.0",
	}
	_, err := helm.LatestVersion()
	if err == nil {
		t.Errorf("Should have failed on bad constraint")
	}
}
