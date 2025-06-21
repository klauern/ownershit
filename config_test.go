package ownershit

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/shurcooL/githubv4"
)

func generateDefaultPermissionsSettings() *PermissionsSettings {
	return &PermissionsSettings{
		BranchPermissions: BranchPermissions{
			AllowSquashMerge: boolPtr(false),
			AllowMergeCommit: boolPtr(false),
			AllowRebaseMerge: boolPtr(false),
		},
		Organization: stringPtr("klauern"),
		Repositories: []*Repository{
			{
				Name:     stringPtr("test"),
				Issues:   boolPtr(false),
				Projects: boolPtr(false),
				Wiki:     boolPtr(false),
			},
		},
		TeamPermissions: []*Permissions{
			{
				Team:  stringPtr("klauern"),
				Level: stringPtr(string(Admin)),
			},
		},
	}
}

func TestMapPermissions(t *testing.T) {
	mocks := setupMocks(t)

	mocks.teamMock.EXPECT().AddTeamRepoBySlug(
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any()).AnyTimes().Return(defaultGoodResponse, nil)

	mocks.graphMock.EXPECT().
		Mutate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes().
		Return(errors.New("dummy error"))

	mocks.repoMock.
		EXPECT().
		Edit(
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
			gomock.Any()).
		Return(nil, defaultGoodResponse, nil)

	mocks.graphMock.EXPECT().Query(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil).
		Do(func(ctx context.Context, query *GetRepoQuery, things map[string]interface{}) {
			query.Repository.ID = githubv4.ID("12345")
			query.Repository.HasIssuesEnabled = false
			query.Repository.HasProjectsEnabled = false
			query.Repository.HasWikiEnabled = false
		})

	type args struct {
		settings *PermissionsSettings
		client   *GitHubClient
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test nothing",
			args: args{
				settings: generateDefaultPermissionsSettings(),
				client:   mocks.client,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			MapPermissions(tt.args.settings, tt.args.client)
		})
	}
}

func TestUpdateBranchMergeStrategies(t *testing.T) {
	mocks := setupMocks(t)

	mocks.teamMock.EXPECT().AddTeamRepoBySlug(
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any()).AnyTimes().Return(defaultGoodResponse, nil)

	mocks.graphMock.EXPECT().
		Mutate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes().
		Return(errors.New("dummy error"))

	mocks.repoMock.
		EXPECT().
		Edit(
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
			gomock.Any()).
		Return(nil, defaultGoodResponse, nil)

	mocks.graphMock.EXPECT().Query(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil).
		Do(func(ctx context.Context, query *GetRepoQuery, things map[string]interface{}) {
			query.Repository.ID = githubv4.ID("12345")
			query.Repository.HasIssuesEnabled = false
			query.Repository.HasProjectsEnabled = false
			query.Repository.HasWikiEnabled = false
		})

	type args struct {
		settings *PermissionsSettings
		client   *GitHubClient
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "default run",
			args: args{
				settings: generateDefaultPermissionsSettings(),
				client:   mocks.client,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			UpdateBranchMergeStrategies(tt.args.settings, tt.args.client)
		})
	}
}

// func TestSyncLabels(t *testing.T) {
// 	mocks := setupMocks(t)
// 	mocks.issuesMock.EXPECT().ListLabels(gomock.Any(), gomock.Any(),
// 		gomock.Any()).Return()
// }
