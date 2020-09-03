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

func TestAMIHappyPath(t *testing.T) {
	ami := AMI{
		Owner: "amazon",
		Name:  "amazon-eks-node-1.13-*",
	}
	latestVersion, err := ami.LatestVersion()
	if err != nil {
		t.Errorf("Failed AMI happy path test: %v", err)
	}
	if latestVersion == "" {
		t.Errorf("Got an empty latestVersion")
	}
}

func TestAMIDoesntExist(t *testing.T) {
	fakeAmi := "this-ami-doesnt-exist-zeitgeist"
	ami := AMI{
		Owner: "amazon",
		Name:  fakeAmi,
	}
	_, err := ami.LatestVersion()
	if err == nil {
		t.Errorf("Found a latest version for unknown AMI: %s", fakeAmi)
	}
}
