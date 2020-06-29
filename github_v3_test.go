package ownershit

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-github/v32/github"
	"github.com/shurcooL/githubv4"
	"github.com/stretchr/testify/assert"

	"github.com/klauern/ownershit/mocks"
)

func stringPtr(s string) *string {
	return &s
}

func int64Ptr(i int64) *int64 {
	return &i
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
					ID:    int64Ptr(0),
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
			if err := c.AddPermissions(&tt.args.organization, &tt.args.repo, &tt.args.perm); (err != nil) != tt.wantErr {
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
