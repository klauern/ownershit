package ownershit

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/go-github/v32/github"
	"github.com/shurcooL/githubv4"
)

func TestGitHubClient_SetRepository(t *testing.T) {
	type fields struct {
		Teams   TeamsService
		Graph   GraphQLClient
		V3      *github.Client
		V4      *githubv4.Client
		Context context.Context
	}
	type args struct {
		id      githubv4.ID
		wiki    bool
		issues  bool
		project bool
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
			if err := c.SetRepository(tt.args.id, tt.args.wiki, tt.args.issues, tt.args.project); (err != nil) != tt.wantErr {
				t.Errorf("GitHubClient.SetRepository() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGitHubClient_SetBranchRules(t *testing.T) {
	type fields struct {
		Teams   TeamsService
		Graph   GraphQLClient
		V3      *github.Client
		V4      *githubv4.Client
		Context context.Context
	}
	type args struct {
		id                       *githubv4.ID
		branchPattern            string
		approverCount            int
		requireCodeOwners        bool
		requiresApprovingReviews bool
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
			if err := c.
				SetBranchRules(
					tt.args.id,
					tt.args.branchPattern,
					tt.args.approverCount,
					tt.args.requireCodeOwners,
					tt.args.requiresApprovingReviews); (err != nil) != tt.wantErr {
				t.Errorf("GitHubClient.SetBranchRules() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGitHubClient_GetRepository(t *testing.T) {
	type fields struct {
		Teams   TeamsService
		Graph   GraphQLClient
		V3      *github.Client
		V4      *githubv4.Client
		Context context.Context
	}
	type args struct {
		name  string
		owner string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *githubv4.ID
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
			got, err := c.GetRepository(tt.args.name, tt.args.owner)
			if (err != nil) != tt.wantErr {
				t.Errorf("GitHubClient.GetRepository() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GitHubClient.GetRepository() = %v, want %v", got, tt.want)
			}
		})
	}
}
