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

package container_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"sigs.k8s.io/zeitgeist/pkg/container"
	"sigs.k8s.io/zeitgeist/pkg/container/containerfakes"
)

func newSUT() (*container.Container, *containerfakes.FakeClient) {
	client := &containerfakes.FakeClient{}
	sut := container.New()
	sut.SetClient(client)

	return sut, client
}

func TestTagsSuccessEmpty(t *testing.T) {
	// Given
	sut, client := newSUT()
	client.ListTagsReturns([]string{}, nil)

	// When
	res, err := sut.Client().ListTags("honk/honk")

	// Then
	require.NoError(t, err)
	require.Empty(t, res)
}

func TestTagsFailed(t *testing.T) {
	// Given
	sut, client := newSUT()
	client.ListTagsReturns([]string{}, errors.New("error"))

	// When
	_, err := sut.Client().ListTags("honk/honk")

	// Then
	require.Error(t, err)
}

func TestTagsSuccess(t *testing.T) {
	// Given
	sut, client := newSUT()
	client.ListTagsReturns([]string{"v1.0.0", "v0.8.0", "v2.0.1"}, nil)

	// When
	res, err := sut.Client().ListTags("honk/honk")

	// Then
	require.NoError(t, err)
	require.Len(t, res, 3)
}
