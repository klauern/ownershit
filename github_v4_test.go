package ownershit

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-github/v43/github"
	"github.com/shurcooL/githubv4"
)

func TestGitHubClient_SetRepository(t *testing.T) {
	mock := setupMocks(t)
	mock.graphMock.EXPECT().Mutate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	mock.graphMock.EXPECT().Mutate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("test error"))

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
	mock.graphMock.EXPECT().Mutate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("forced expected error"))

	err := mock.client.SetBranchRules(githubv4.ID("test"), github.String("master"), github.Int(1), github.Bool(true), github.Bool(true))
	if err != nil {
		t.Errorf("GitHubClient.SetBranchRuels() error = %v, wantErr %v", err, false)
	}
	err = mock.client.SetBranchRules(githubv4.ID("test"), github.String("master"), github.Int(1), github.Bool(true), github.Bool(true))
	if err == nil {
		t.Errorf("GitHubClient.SetBranchRuels() error = %v, wantErr %v", err, true)
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
		Return(errors.New("forced expected error")).
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
