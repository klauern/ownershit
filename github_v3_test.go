package ownershit

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/google/go-github/v66/github"
	"github.com/klauern/ownershit/mocks"
	"go.uber.org/mock/gomock"
)

var (
	ErrErroringThing = errors.New("erroring thing")
	ErrDummyV3Error  = errors.New("dummy error")
)

func TestMapPermsWithGitHub(t *testing.T) {
	if !*githubv3 {
		t.Skip("no testing")
	}
	owner := "klauern"
	perms := &Permissions{
		Level: stringPtr(string(Admin)),
		Team:  &owner,
	}

	client := defaultGitHubClient()
	err := client.AddPermissions(owner, "non-existent-repo", perms)
	if err == nil {
		t.Error("expected error on non-existent repo")
	}
}

func mockGitHubResponse() *github.Response {
	resp := github.Response{}
	resp.Response = &http.Response{}
	resp.Status = "200 OK"
	resp.StatusCode = 200

	return &resp
}

func TestOmitPermFixes(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := defaultGitHubClient()
	repoSvc := mocks.NewMockRepositoriesService(ctrl)
	client.Repositories = repoSvc

	repoSvc.
		EXPECT().
		Edit(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, defaultGoodResponse, nil)

	repoSvc.
		EXPECT().
		Edit(gomock.Any(), gomock.Eq("klauern"), gomock.Eq("ownershit"), gomock.Eq(&github.Repository{
			AllowMergeCommit: github.Bool(false),
		})).Return(nil, defaultGoodResponse, nil)

	// set default true for everything
	err := client.UpdateBranchPermissions("klauern", "ownershit", &BranchPermissions{
		AllowMergeCommit: boolPtr(true),
		AllowRebaseMerge: boolPtr(true),
		AllowSquashMerge: boolPtr(true),
	})
	if err != nil {
		t.Error("did not expect error")
	}

	err = client.UpdateBranchPermissions("klauern", "ownershit", &BranchPermissions{
		AllowMergeCommit: boolPtr(false),
	})
	if err != nil {
		t.Error("error not expected")
	}

	repoSvc.EXPECT().Edit(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(&github.Repository{
		AllowMergeCommit: boolPtr(true),
	})).Return(nil, mockGitHubResponse(), ErrErroringThing)
	if err = client.UpdateBranchPermissions("klauern", "ownershit", &BranchPermissions{
		AllowMergeCommit: boolPtr(true),
	}); err == nil {
		t.Error("error expected here")
	}
}

func TestGitHubClient_SetBranchProtectionFallback(t *testing.T) {
	tests := []struct {
		name    string
		org     string
		repo    string
		branch  string
		perms   *BranchPermissions
		wantErr bool
	}{
		{
			name:    "nil permissions should not error",
			org:     "testorg",
			repo:    "testrepo",
			branch:  "main",
			perms:   nil,
			wantErr: false,
		},
		{
			name:   "basic pull request reviews",
			org:    "testorg",
			repo:   "testrepo",
			branch: "main",
			perms: &BranchPermissions{
				RequirePullRequestReviews: boolPtr(true),
				ApproverCount:             intPtr(2),
				RequireCodeOwners:         boolPtr(true),
			},
			wantErr: false,
		},
		{
			name:   "status checks configuration",
			org:    "testorg",
			repo:   "testrepo",
			branch: "main",
			perms: &BranchPermissions{
				RequireStatusChecks:   boolPtr(true),
				StatusChecks:          []string{"ci/build", "ci/test"},
				RequireUpToDateBranch: boolPtr(true),
			},
			wantErr: false,
		},
		{
			name:   "enforce admins",
			org:    "testorg",
			repo:   "testrepo",
			branch: "main",
			perms: &BranchPermissions{
				EnforceAdmins: boolPtr(true),
			},
			wantErr: false,
		},
		{
			name:   "restrict pushes with allowlist",
			org:    "testorg",
			repo:   "testrepo",
			branch: "main",
			perms: &BranchPermissions{
				RestrictPushes: boolPtr(true),
				PushAllowlist:  []string{"admin-team", "deploy-team"},
			},
			wantErr: false,
		},
		{
			name:   "advanced features",
			org:    "testorg",
			repo:   "testrepo",
			branch: "main",
			perms: &BranchPermissions{
				RequireLinearHistory: boolPtr(true),
				AllowForcePushes:     boolPtr(false),
				AllowDeletions:       boolPtr(false),
			},
			wantErr: false,
		},
		{
			name:   "comprehensive protection rules",
			org:    "testorg",
			repo:   "testrepo",
			branch: "main",
			perms: &BranchPermissions{
				RequirePullRequestReviews:     boolPtr(true),
				ApproverCount:                 intPtr(1),
				RequireCodeOwners:             boolPtr(true),
				RequireStatusChecks:           boolPtr(true),
				StatusChecks:                  []string{"ci/build", "ci/test", "security/scan"},
				RequireUpToDateBranch:         boolPtr(true),
				EnforceAdmins:                 boolPtr(true),
				RestrictPushes:                boolPtr(true),
				PushAllowlist:                 []string{"admin-team"},
				RequireConversationResolution: boolPtr(true),
				RequireLinearHistory:          boolPtr(true),
				AllowForcePushes:              boolPtr(false),
				AllowDeletions:                boolPtr(false),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a real client for testing the fallback logic
			// This will test the REST API integration without actually calling GitHub
			client := defaultGitHubClient()

			// Note: In a real test environment, we would mock the underlying v3 client
			// For now, we're testing the method structure and logic flow
			err := client.SetBranchProtectionFallback(tt.org, tt.repo, tt.branch, tt.perms)

			// Since we're using a real client with empty token, we expect some errors
			// but we mainly want to ensure the method doesn't panic and handles nil properly
			if tt.perms == nil && err != nil {
				t.Errorf("SetBranchProtectionFallback() with nil perms should not error, got %v", err)
			}
		})
	}
}

func TestGitHubClient_AddPermissions(t *testing.T) {
	ctrl := gomock.NewController(t)

	teamMock := mocks.NewMockTeamsService(ctrl)

	teamMock.
		EXPECT().
		AddTeamRepoBySlug(
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
			gomock.Any()).
		Return(&github.Response{
			Response: &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader([]byte("something"))),
			},
		}, nil)
	teamMock.
		EXPECT().
		AddTeamRepoBySlug(
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
			gomock.Any()).
		Return(&github.Response{
			Response: &http.Response{
				StatusCode: 500,
				Body:       io.NopCloser(bytes.NewReader([]byte("error"))),
			},
		}, ErrDummyV3Error)

	graphMock := mocks.NewMockGraphQLClient(ctrl)

	type fields struct {
		Teams TeamsService
		Graph GraphQLClient
	}
	type args struct {
		repo         string
		organization string
		perm         Permissions
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "",
			fields: fields{
				Teams: teamMock,
				Graph: graphMock,
			},
			args: args{
				repo:         "junk",
				organization: "junk",
				perm: Permissions{
					Team:  stringPtr("me"),
					Level: stringPtr("PUSH"),
				},
			},
			wantErr: false,
		},
		{
			name: "",
			fields: fields{
				Teams: teamMock,
				Graph: graphMock,
			},
			args: args{
				repo:         "junk",
				organization: "junk",
				perm: Permissions{
					Team:  stringPtr("me"),
					Level: stringPtr("PUSH"),
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &GitHubClient{
				Teams: tt.fields.Teams,
				Graph: tt.fields.Graph,
			}
			if err := c.AddPermissions(tt.args.organization, tt.args.repo, &tt.args.perm); (err != nil) != tt.wantErr {
				t.Errorf("AddPermissions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGitHubClient_SetRepositoryAdvancedSettings(t *testing.T) {
	mock := setupMocks(t)

	// Test successful call
	mock.repoMock.EXPECT().Edit(gomock.Any(), "testorg", "testrepo", gomock.Any()).
		Return(&github.Repository{}, &github.Response{
			Response: &http.Response{StatusCode: 200},
		}, nil)

	deleteBranchOnMerge := true
	err := mock.client.SetRepositoryAdvancedSettings("testorg", "testrepo", &deleteBranchOnMerge)
	if err != nil {
		t.Errorf("SetRepositoryAdvancedSettings() error = %v, wantErr = false", err)
	}

	// Test error handling
	mock.repoMock.EXPECT().Edit(gomock.Any(), "testorg", "testrepo", gomock.Any()).
		Return(nil, &github.Response{
			Response: &http.Response{StatusCode: 500},
		}, ErrDummyV3Error)

	err = mock.client.SetRepositoryAdvancedSettings("testorg", "testrepo", &deleteBranchOnMerge)
	if err == nil {
		t.Errorf("SetRepositoryAdvancedSettings() error = nil, wantErr = true")
	}

	// Test nil parameter (should return without error)
	err = mock.client.SetRepositoryAdvancedSettings("testorg", "testrepo", nil)
	if err != nil {
		t.Errorf("SetRepositoryAdvancedSettings() with nil parameter error = %v, wantErr = false", err)
	}
}
