/*
Copyright 2021 The Kubernetes Authors.

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

package upstream

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestUnserialiseContainer(t *testing.T) {
	validYamls := []string{
		"flavour: container\nurl: honk/honk\nconstraints: <1.0.0",
	}

	for _, valid := range validYamls {
		var u Container

		err := yaml.Unmarshal([]byte(valid), &u)
		if err != nil {
			t.Errorf("Failed to deserialise valid yaml:\n%s", valid)
		}
	}
}
