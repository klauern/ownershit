package ownershit

import (
	"bytes"
	"context"
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-github/v32/github"
	"github.com/klauern/ownershit/mocks"
	"github.com/rs/zerolog"
	"github.com/shurcooL/githubv4"
	"github.com/stretchr/testify/assert"
)

func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func defaultGitHubClient() *GitHubClient {
	return NewGitHubClient(context.TODO(), "")
}

var (
	githubv3 = flag.Bool("githubv3", false, "run GitHub V3 Integration Tests")
)

func TestMain(m *testing.M) {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	flag.Parse()
	result := m.Run()
	os.Exit(result)
}

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

	defaultGoodResponse := &github.Response{
		Response: &http.Response{
			StatusCode: 200,
		},
	}

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

	graphMock := mocks.NewMockGraphQLClient(ctrl)

	type fields struct {
		Teams   TeamsService
		Graph   GraphQLClient
		V3      *github.Client
		V4      *githubv4.Client
		Context context.Context
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
				Teams:   teamMock,
				Graph:   graphMock,
				V3:      nil,
				V4:      nil,
				Context: nil,
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &GitHubClient{
				Teams:   tt.fields.Teams,
				Graph:   tt.fields.Graph,
				V3:      tt.fields.V3,
				V4:      tt.fields.V4,
				Context: tt.fields.Context,
			}
			if err := c.AddPermissions(tt.args.organization, tt.args.repo, &tt.args.perm); (err != nil) != tt.wantErr {
				t.Errorf("AddPermissions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGitHubClient_AddPermissions2(t *testing.T) {
	ctrl := gomock.NewController(t)
	teamMock := mocks.NewMockTeamsService(ctrl)

	any := gomock.Any()

	teamMock.
		EXPECT().
		AddTeamRepoBySlug(
			any,
			any,
			any,
			"klauern",
			"ownershit",
			&github.TeamAddTeamRepoOptions{Permission: "push"},
		).
		Return(&github.Response{
			Response: &http.Response{
				StatusCode: 0,
			},
		}, nil)

	client := &GitHubClient{
		Teams:   teamMock,
		Graph:   nil,
		V3:      nil,
		V4:      nil,
		Context: nil,
	}

	_, err := client.Teams.AddTeamRepoBySlug(context.Background(), "", "", "klauern", "ownershit", &github.TeamAddTeamRepoOptions{
		Permission: "push",
	})
	if err != nil {
		t.Error(err)
	}

	assert.NoError(t, err)
}
