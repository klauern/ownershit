package ownershit

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-github/v32/github"
	"github.com/klauern/ownershit/mocks"
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
	err := client.UpdateRepositorySettings("klauern", "ownershit", &BranchPermissions{
		AllowMergeCommit: boolPtr(true),
		AllowRebaseMerge: boolPtr(true),
		AllowSquashMerge: boolPtr(true),
	})
	if err != nil {
		t.Error("did not expect error")
	}

	err = client.UpdateRepositorySettings("klauern", "ownershit", &BranchPermissions{
		AllowMergeCommit: boolPtr(false),
	})
	if err != nil {
		t.Error("error not expected")
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
				Body:       ioutil.NopCloser(bytes.NewReader([]byte("something"))),
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
				Body:       ioutil.NopCloser(bytes.NewReader([]byte("error"))),
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
