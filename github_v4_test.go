package ownershit

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-github/v66/github"
	"github.com/shurcooL/githubv4"
)

var (
	ErrTestV4Error         = errors.New("test error")
	ErrForcedExpectedError = errors.New("forced expected error")
)

func TestGitHubClient_SetRepository(t *testing.T) {
	mock := setupMocks(t)
	mock.graphMock.EXPECT().Mutate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	mock.graphMock.EXPECT().Mutate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(ErrTestV4Error)

	if err := mock.client.SetRepository(githubv4.ID("test"), boolPtr(false), boolPtr(false), boolPtr(false)); (err != nil) != false {
		t.Errorf("GitHubClient.SetRepository() error = %v, wantErr %v", err, false)
	}
	if err := mock.client.SetRepository(githubv4.ID("test"), boolPtr(false), boolPtr(false), boolPtr(false)); (err != nil) != true {
		t.Errorf("GitHubClient.SetRepository() error = %v, wantErr %v", err, true)
	}
}

func TestGitHubClient_SetBranchRules(t *testing.T) {
	mock := setupMocks(t)
	mock.graphMock.EXPECT().Mutate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	mock.graphMock.EXPECT().Mutate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(ErrForcedExpectedError)

	err := mock.client.SetBranchRules(githubv4.ID("test"), github.String("master"), github.Int(1), github.Bool(true), github.Bool(true))
	if err != nil {
		t.Errorf("GitHubClient.SetBranchRuels() error = %v, wantErr %v", err, false)
	}
	err = mock.client.SetBranchRules(githubv4.ID("test"), github.String("master"), github.Int(1), github.Bool(true), github.Bool(true))
	if err == nil {
		t.Errorf("GitHubClient.SetBranchRuels() error = %v, wantErr %v", err, true)
	}
}

func TestGitHubClient_SetEnhancedBranchProtection(t *testing.T) {
	mock := setupMocks(t)

	tests := []struct {
		name          string
		repoID        githubv4.ID
		branchPattern string
		perms         *BranchPermissions
		wantErr       bool
		mockSetup     func()
	}{
		{
			name:          "basic protection rules",
			repoID:        githubv4.ID("repo123"),
			branchPattern: "main",
			perms: &BranchPermissions{
				RequirePullRequestReviews: boolPtr(true),
				ApproverCount:             intPtr(2),
				RequireCodeOwners:         boolPtr(true),
			},
			wantErr: false,
			mockSetup: func() {
				mock.graphMock.EXPECT().Mutate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			name:          "advanced status checks",
			repoID:        githubv4.ID("repo456"),
			branchPattern: "main",
			perms: &BranchPermissions{
				RequireStatusChecks:   boolPtr(true),
				StatusChecks:          []string{"ci/build", "ci/test", "security/scan"},
				RequireUpToDateBranch: boolPtr(true),
			},
			wantErr: false,
			mockSetup: func() {
				mock.graphMock.EXPECT().Mutate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			name:          "comprehensive protection with advanced features",
			repoID:        githubv4.ID("repo789"),
			branchPattern: "main",
			perms: &BranchPermissions{
				RequirePullRequestReviews:     boolPtr(true),
				ApproverCount:                 intPtr(1),
				RequireCodeOwners:             boolPtr(true),
				RequireStatusChecks:           boolPtr(true),
				StatusChecks:                  []string{"ci/build"},
				RequireUpToDateBranch:         boolPtr(true),
				EnforceAdmins:                 boolPtr(true),
				RestrictPushes:                boolPtr(true),
				PushAllowlist:                 []string{"admin-team", "deploy-team"},
				RequireConversationResolution: boolPtr(true),
				RequireLinearHistory:          boolPtr(true),
				AllowForcePushes:              boolPtr(false),
				AllowDeletions:                boolPtr(false),
			},
			wantErr: false,
			mockSetup: func() {
				mock.graphMock.EXPECT().Mutate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			name:          "API error handling",
			repoID:        githubv4.ID("repo_error"),
			branchPattern: "main",
			perms: &BranchPermissions{
				RequirePullRequestReviews: boolPtr(true),
			},
			wantErr: true,
			mockSetup: func() {
				mock.graphMock.EXPECT().Mutate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(ErrTestV4Error)
			},
		},
		{
			name:          "large approver count handling",
			repoID:        githubv4.ID("repo_large"),
			branchPattern: "main",
			perms: &BranchPermissions{
				RequirePullRequestReviews: boolPtr(true),
				ApproverCount:             intPtr(2147483648), // Exceeds int32 max
			},
			wantErr: false,
			mockSetup: func() {
				mock.graphMock.EXPECT().Mutate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			err := mock.client.SetEnhancedBranchProtection(tt.repoID, tt.branchPattern, tt.perms)
			if (err != nil) != tt.wantErr {
				t.Errorf("GitHubClient.SetEnhancedBranchProtection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGitHubClient_GetRepository(t *testing.T) {
	mock := setupMocks(t)
	mock.graphMock.
		EXPECT().
		Query(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).
		Do(func(ctx context.Context, query *GetRepoQuery, things map[string]interface{}) {
			query.Repository.ID = githubv4.ID("12345")
			query.Repository.HasIssuesEnabled = false
			query.Repository.HasProjectsEnabled = false
			query.Repository.HasWikiEnabled = false
		})
	mock.graphMock.
		EXPECT().
		Query(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(ErrForcedExpectedError).
		Do(func(ctx context.Context, query *GetRepoQuery, things map[string]interface{}) {
			query.Repository.ID = githubv4.ID("12345")
			query.Repository.HasIssuesEnabled = false
			query.Repository.HasProjectsEnabled = false
			query.Repository.HasWikiEnabled = false
		})

	id, err := mock.client.GetRepository(github.String("ownershit"), github.String("klauern"))
	if err != nil {
		t.Errorf("GitHubClient.GetRepository() error = %v, wantErr %v", err, false)
	}
	if id != githubv4.ID("12345") {
		t.Errorf("GitHubClient.GetRepository() expected id to be %v, got %v", "12345", id)
	}
	id, err = mock.client.GetRepository(github.String("ownershit"), github.String("klauern"))
	if err == nil {
		t.Errorf("GitHubClient.GetRepository() error = %v, wantErr %v", err, true)
	}
	if id != nil {
		t.Errorf("GitHubClient.GetRepository() expected id to be %v, got %v", nil, id)
	}
}
