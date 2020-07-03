package ownershit

import (
	"context"
	"testing"

	"github.com/google/go-github/v32/github"
	"github.com/shurcooL/githubv4"
)

func TestGitHubClient_SetDefaultBranch(t *testing.T) {
	type fields struct {
		Teams   TeamsService
		Graph   GraphQLClient
		V3      *github.Client
		V4      *githubv4.Client
		Context context.Context
	}
	type args struct {
		ctx               context.Context
		owner             string
		repo              string
		defaultBranchName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
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
			if err := c.SetDefaultBranch(
				tt.args.ctx,
				tt.args.owner,
				tt.args.repo,
				tt.args.defaultBranchName); (err != nil) != tt.wantErr {
				t.Errorf("GitHubClient.SetDefaultBranch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGitHubClient_SetRepositoryDefaults(t *testing.T) {
	type fields struct {
		Teams   TeamsService
		Graph   GraphQLClient
		V3      *github.Client
		V4      *githubv4.Client
		Context context.Context
	}
	type args struct {
		ctx      context.Context
		owner    string
		repo     string
		wiki     bool
		issues   bool
		projects bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
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
			if err := c.SetRepositoryDefaults(
				tt.args.ctx,
				tt.args.owner,
				tt.args.repo,
				tt.args.wiki,
				tt.args.issues,
				tt.args.projects); (err != nil) != tt.wantErr {
				t.Errorf("GitHubClient.SetRepositoryDefaults() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
