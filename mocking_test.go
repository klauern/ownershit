package ownershit

import (
	"context"
	"flag"
	"net/http"
	"os"
	"testing"

	"github.com/google/go-github/v66/github"
	"github.com/klauern/ownershit/mocks"
	"github.com/rs/zerolog"
	"go.uber.org/mock/gomock"
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

var defaultGoodResponse = &github.Response{
	Response: &http.Response{
		StatusCode: 200,
	},
}

var githubv3 = flag.Bool("githubv3", false, "run GitHub V3 Integration Tests")

func TestMain(m *testing.M) {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	flag.Parse()
	result := m.Run()
	os.Exit(result)
}

type testMocks struct {
	ctrl       *gomock.Controller
	client     *GitHubClient
	teamMock   *mocks.MockTeamsService
	repoMock   *mocks.MockRepositoriesService
	graphMock  *mocks.MockGraphQLClient
	issuesMock *mocks.MockIssuesService
}

func setupMocks(t *testing.T) *testMocks {
	ctrl := gomock.NewController(t)
	teams := mocks.NewMockTeamsService(ctrl)
	graph := mocks.NewMockGraphQLClient(ctrl)
	repo := mocks.NewMockRepositoriesService(ctrl)
	issues := mocks.NewMockIssuesService(ctrl)

	// Create a real GitHub client for v3 operations that can't be mocked easily
	// Use an empty token since we're mocking the service layer
	realClient := defaultGitHubClient()

	ghClient := &GitHubClient{
		Teams:        teams,
		Context:      context.TODO(),
		Graph:        graph,
		Repositories: repo,
		Issues:       issues,
		v3:           realClient.v3, // Set the v3 client to avoid nil pointer
		v4:           realClient.v4, // Set the v4 client as well for completeness
	}
	return &testMocks{
		ctrl:       ctrl,
		client:     ghClient,
		graphMock:  graph,
		repoMock:   repo,
		teamMock:   teams,
		issuesMock: issues,
	}
}
