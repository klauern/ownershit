package ownershit

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/google/go-github/v66/github"
	"github.com/klauern/ownershit/mocks"
	"go.uber.org/mock/gomock"
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
	})).Return(nil, mockGitHubResponse(), fmt.Errorf("erroring thing"))
	if err = client.UpdateBranchPermissions("klauern", "ownershit", &BranchPermissions{
		AllowMergeCommit: boolPtr(true),
	}); err == nil {
		t.Error("error expected here")
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
		}, errors.New("dummy error"))

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
