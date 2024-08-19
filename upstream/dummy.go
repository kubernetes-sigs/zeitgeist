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

package upstream

// Dummy upstream always returns a fixed latest version, by default 1.0.0. Can be used for testing.
type Dummy struct {
	Base
	Latest string
}

// LatestVersion always returns a fixed version.
func (upstream Dummy) LatestVersion() (string, error) {
	if upstream.Latest != "" {
		return upstream.Latest, nil
	}
	return "1.0.0", nil
}
