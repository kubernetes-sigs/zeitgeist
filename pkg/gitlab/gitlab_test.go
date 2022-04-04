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

package gitlab_test

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	gogitlab "github.com/xanzy/go-gitlab"

	"sigs.k8s.io/zeitgeist/pkg/gitlab"
	"sigs.k8s.io/zeitgeist/pkg/gitlab/gitlabfakes"
)

func newSUT() (*gitlab.GitLab, *gitlabfakes.FakeClient) {
	os.Setenv("GITLAB_TOKEN", "honk")
	client := &gitlabfakes.FakeClient{}
	sut := gitlab.New()
	sut.SetClient(client)

	return sut, client
}

func newSUTPrivate() (*gitlab.GitLab, *gitlabfakes.FakeClient) {
	os.Setenv("GITLAB_PRIVATE_TOKEN", "private_honk")
	client := &gitlabfakes.FakeClient{}
	sut := gitlab.NewPrivate("https://honk.gitlab.com/")
	sut.SetClient(client)

	return sut, client
}

func TestBranchesSuccessEmpty(t *testing.T) {
	// Given
	sut, client := newSUT()
	client.ListBranchesReturns([]*gogitlab.Branch{}, nil, nil)

	// When
	res, err := sut.Branches("", "")

	// Then
	require.Nil(t, err)
	require.Empty(t, res)
}

func TestPrivateBranchesSuccessEmpty(t *testing.T) {
	// Given
	sut, client := newSUTPrivate()
	client.ListBranchesReturns([]*gogitlab.Branch{}, nil, nil)

	// When
	res, err := sut.Branches("", "")

	// Then
	require.Nil(t, err)
	require.Empty(t, res)
}

func TestBranchesFailed(t *testing.T) {
	// Given
	sut, client := newSUT()
	client.ListBranchesReturns(nil, nil, errors.New("error"))

	// When
	res, err := sut.Branches("", "")

	// Then
	require.NotNil(t, err)
	require.Nil(t, res, nil)
}

func TestPrivateBranchesFailed(t *testing.T) {
	// Given
	sut, client := newSUTPrivate()
	client.ListBranchesReturns(nil, nil, errors.New("error"))

	// When
	res, err := sut.Branches("", "")

	// Then
	require.NotNil(t, err)
	require.Nil(t, res, nil)
}

func TestReleasesSuccessEmpty(t *testing.T) {
	// Given
	sut, client := newSUT()
	client.ListReleasesReturns([]*gogitlab.Release{}, nil, nil)

	// When
	res, err := sut.Releases("", "")

	// Then
	require.Nil(t, err)
	require.Empty(t, res)
}

func TestPrivateReleasesSuccessEmpty(t *testing.T) {
	// Given
	sut, client := newSUTPrivate()
	client.ListReleasesReturns([]*gogitlab.Release{}, nil, nil)

	// When
	res, err := sut.Releases("", "")

	// Then
	require.Nil(t, err)
	require.Empty(t, res)
}

func TestReleasesFailed(t *testing.T) {
	// Given
	sut, client := newSUT()
	client.ListReleasesReturns(nil, nil, errors.New("error"))

	// When
	res, err := sut.Releases("", "")

	// Then
	require.NotNil(t, err)
	require.Nil(t, res, nil)
}

func TestPrivateReleasesFailed(t *testing.T) {
	// Given
	sut, client := newSUTPrivate()
	client.ListReleasesReturns(nil, nil, errors.New("error"))

	// When
	res, err := sut.Releases("", "")

	// Then
	require.NotNil(t, err)
	require.Nil(t, res, nil)
}

func TestReleasesSuccessNoPreReleases(t *testing.T) {
	// Given
	var (
		tag1 = "v1.18.0"
		tag2 = "v1.17.0"
		tag3 = "v1.16.0"
	)
	sut, client := newSUT()
	client.ListReleasesReturns([]*gogitlab.Release{
		{TagName: tag1},
		{TagName: tag2},
		{TagName: tag3},
	}, nil, nil)

	// When
	res, err := sut.Releases("", "")

	// Then
	require.Nil(t, err)
	require.Len(t, res, 3)
	require.Equal(t, tag1, res[0].TagName)
	require.Equal(t, tag2, res[1].TagName)
	require.Equal(t, tag3, res[2].TagName)
}

func TestListProjects(t *testing.T) {
	// Given
	sut, client := newSUT()
	client.ListProjectsReturns([]*gogitlab.Project{
		{ID: 1, Name: "honk"},
	}, nil, nil)

	// When
	res, err := sut.GetRepository("honkcorp", "honk")
	// Then
	require.NoError(t, err)
	require.Equal(t, "honk", res.Name)
}

func TestListProjectsNoProjects(t *testing.T) {
	// Given
	sut, client := newSUT()
	client.ListProjectsReturns([]*gogitlab.Project{}, nil, nil)

	// When
	_, err := sut.GetRepository("honkcorp", "honk")
	// Then
	require.Error(t, err)
	require.EqualError(t, err, "no project found")
}

func TestListProjectsMoreProjects(t *testing.T) {
	// Given
	sut, client := newSUT()
	client.ListProjectsReturns([]*gogitlab.Project{
		{ID: 1, Name: "honk"},
		{ID: 3, Name: "honk"},
	}, nil, nil)

	// When
	_, err := sut.GetRepository("honkcorp", "honk")
	// Then
	require.Error(t, err)
	require.EqualError(t, err, "expected one project got 2")
}

func TestListTags(t *testing.T) {
	// Given
	var (
		tag1 = "v1.18.0"
		tag2 = "v1.17.0"
		tag3 = "v1.16.0"
	)
	sut, client := newSUT()
	client.ListTagsReturns([]*gogitlab.Tag{
		{Name: tag1},
		{Name: tag2},
		{Name: tag3},
	}, nil, nil)

	// When
	res, err := sut.ListTags("", "")

	// Then
	require.Nil(t, err)
	require.Len(t, res, 3)
	require.Equal(t, tag1, res[0].Name)
	require.Equal(t, tag2, res[1].Name)
	require.Equal(t, tag3, res[2].Name)
}
