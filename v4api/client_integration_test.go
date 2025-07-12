package v4api

import (
	"context"
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/Khan/genqlient/graphql"
	"go.uber.org/mock/gomock"

	mock_graphql "github.com/klauern/ownershit/v4api/mocks"
)

const (
	getRateLimitMethod = "GetRateLimit"
	syncLabelsMethod   = "SyncLabels"
)

func TestGitHubV4Client_GetTeams_PaginationScenarios(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name          string
		mockSetup     func(*mock_graphql.MockClient)
		expectPanic   bool
		expectError   bool
		expectedCalls int
	}{
		{
			name: "single page of teams",
			mockSetup: func(mockClient *mock_graphql.MockClient) {
				mockClient.EXPECT().MakeRequest(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, req *graphql.Request, resp *graphql.Response) error {
						// Set up a response with no pagination
						if data, ok := resp.Data.(*GetTeamsResponse); ok {
							data.Organization = GetTeamsOrganization{
								Teams: GetTeamsOrganizationTeamsTeamConnection{
									PageInfo: GetTeamsOrganizationTeamsTeamConnectionPageInfo{
										HasNextPage: false,
										StartCursor: "",
										EndCursor:   "",
									},
									Edges: []GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdge{
										{Node: GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeam{
											Name:        "test-team",
											Description: "A test team",
											Members: GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamMembersTeamMemberConnection{
												Edges: []GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamMembersTeamMemberConnectionEdgesTeamMemberEdge{},
											},
											ChildTeams: GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamChildTeamsTeamConnection{
												Edges: []GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamChildTeamsTeamConnectionEdgesTeamEdge{},
											},
										}},
									},
								},
							}
						}
						return nil
					},
				).Times(1)
			},
			expectPanic:   false,
			expectError:   false,
			expectedCalls: 1,
		},
		{
			name: "multiple pages of teams",
			mockSetup: func(mockClient *mock_graphql.MockClient) {
				// First call returns hasNextPage: true
				first := mockClient.EXPECT().MakeRequest(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, req *graphql.Request, resp *graphql.Response) error {
						if data, ok := resp.Data.(*GetTeamsResponse); ok {
							data.Organization = GetTeamsOrganization{
								Teams: GetTeamsOrganizationTeamsTeamConnection{
									PageInfo: GetTeamsOrganizationTeamsTeamConnectionPageInfo{
										HasNextPage: true,
										StartCursor: "cursor1",
										EndCursor:   "cursor2",
									},
									Edges: []GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdge{
										{Node: GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeam{
											Name:        "team1",
											Description: "First team",
											Members: GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamMembersTeamMemberConnection{
												Edges: []GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamMembersTeamMemberConnectionEdgesTeamMemberEdge{},
											},
											ChildTeams: GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamChildTeamsTeamConnection{
												Edges: []GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamChildTeamsTeamConnectionEdgesTeamEdge{},
											},
										}},
									},
								},
							}
						}
						return nil
					},
				).Times(1)
				// Second call returns hasNextPage: false
				mockClient.EXPECT().MakeRequest(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, req *graphql.Request, resp *graphql.Response) error {
						if data, ok := resp.Data.(*GetTeamsResponse); ok {
							data.Organization = GetTeamsOrganization{
								Teams: GetTeamsOrganizationTeamsTeamConnection{
									PageInfo: GetTeamsOrganizationTeamsTeamConnectionPageInfo{
										HasNextPage: false,
										StartCursor: "cursor2",
										EndCursor:   "cursor3",
									},
									Edges: []GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdge{
										{Node: GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeam{
											Name:        "team2",
											Description: "Second team",
											Members: GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamMembersTeamMemberConnection{
												Edges: []GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamMembersTeamMemberConnectionEdgesTeamMemberEdge{},
											},
											ChildTeams: GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamChildTeamsTeamConnection{
												Edges: []GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamChildTeamsTeamConnectionEdgesTeamEdge{},
											},
										}},
									},
								},
							}
						}
						return nil
					},
				).Times(1).After(first)
			},
			expectPanic:   false,
			expectError:   false,
			expectedCalls: 2,
		},
		{
			name: "error on first call",
			mockSetup: func(mockClient *mock_graphql.MockClient) {
				mockClient.EXPECT().MakeRequest(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("api error")).Times(1)
			},
			expectPanic:   true,
			expectError:   false,
			expectedCalls: 1,
		},
		{
			name: "error on subsequent call",
			mockSetup: func(mockClient *mock_graphql.MockClient) {
				// First call succeeds with hasNextPage: true
				first := mockClient.EXPECT().MakeRequest(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, req *graphql.Request, resp *graphql.Response) error {
						if data, ok := resp.Data.(*GetTeamsResponse); ok {
							data.Organization = GetTeamsOrganization{
								Teams: GetTeamsOrganizationTeamsTeamConnection{
									PageInfo: GetTeamsOrganizationTeamsTeamConnectionPageInfo{
										HasNextPage: true,
										StartCursor: "cursor1",
										EndCursor:   "cursor2",
									},
									Edges: []GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdge{
										{Node: GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeam{
											Name:        "team1",
											Description: "First team",
											Members: GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamMembersTeamMemberConnection{
												Edges: []GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamMembersTeamMemberConnectionEdgesTeamMemberEdge{},
											},
											ChildTeams: GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamChildTeamsTeamConnection{
												Edges: []GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamChildTeamsTeamConnectionEdgesTeamEdge{},
											},
										}},
									},
								},
							}
						}
						return nil
					},
				).Times(1)
				// Second call fails
				mockClient.EXPECT().MakeRequest(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("pagination error")).Times(1).After(first)
			},
			expectPanic:   true,
			expectError:   false,
			expectedCalls: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := mock_graphql.NewMockClient(ctrl)
			ctx := context.Background()

			client := &GitHubV4Client{
				Context: ctx,
				client:  mockClient,
			}

			tt.mockSetup(mockClient)

			teams, err := client.GetTeams("test-org")

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if tt.expectPanic && err == nil {
				t.Error("Expected error (instead of panic) but got none")
			}

			if !tt.expectError && !tt.expectPanic && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectPanic && teams == nil {
				t.Error("Expected teams but got nil")
			}
		})
	}
}

func TestGitHubV4Client_GetRateLimit_ErrorHandling(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name          string
		mockSetup     func(*mock_graphql.MockClient)
		expectError   bool
		errorContains string
	}{
		{
			name: "successful rate limit fetch",
			mockSetup: func(mockClient *mock_graphql.MockClient) {
				mockClient.EXPECT().MakeRequest(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			},
			expectError: false,
		},
		{
			name: "GraphQL error",
			mockSetup: func(mockClient *mock_graphql.MockClient) {
				mockClient.EXPECT().MakeRequest(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("graphql error")).Times(1)
			},
			expectError:   true,
			errorContains: "can't get ratelimit",
		},
		{
			name: "network error",
			mockSetup: func(mockClient *mock_graphql.MockClient) {
				mockClient.EXPECT().MakeRequest(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("network timeout")).Times(1)
			},
			expectError:   true,
			errorContains: "can't get ratelimit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := mock_graphql.NewMockClient(ctrl)
			ctx := context.Background()

			client := &GitHubV4Client{
				Context: ctx,
				client:  mockClient,
			}

			tt.mockSetup(mockClient)

			rateLimit, err := client.GetRateLimit()

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.errorContains != "" && !containsString(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorContains, err)
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			} else {
				// RateLimit should be valid (empty struct is valid for this test)
				_ = rateLimit
			}
		})
	}
}

func TestGitHubV4Client_SyncLabels_ComplexScenarios(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name          string
		repo          string
		labels        []Label
		mockSetup     func(*mock_graphql.MockClient)
		expectError   bool
		errorContains string
	}{
		{
			name:   "successful sync with empty labels",
			repo:   "test-repo",
			labels: []Label{},
			mockSetup: func(mockClient *mock_graphql.MockClient) {
				// Mock GetRepositoryIssueLabels call - returns empty labels
				mockClient.EXPECT().MakeRequest(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, req *graphql.Request, resp *graphql.Response) error {
						if data, ok := resp.Data.(*GetRepositoryIssueLabelsResponse); ok {
							data.Repository = GetRepositoryIssueLabelsRepository{
								Id: "repo123",
								Labels: GetRepositoryIssueLabelsRepositoryLabelsLabelConnection{
									Edges: []GetRepositoryIssueLabelsRepositoryLabelsLabelConnectionEdgesLabelEdge{},
									PageInfo: GetRepositoryIssueLabelsRepositoryLabelsLabelConnectionPageInfo{
										HasNextPage: false,
										EndCursor:   "",
										StartCursor: "",
									},
								},
							}
						}
						return nil
					},
				).Times(1)
			},
			expectError: false,
		},
		{
			name:   "error getting repository labels",
			repo:   "test-repo",
			labels: []Label{},
			mockSetup: func(mockClient *mock_graphql.MockClient) {
				mockClient.EXPECT().MakeRequest(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("repo not found")).Times(1)
			},
			expectError:   true,
			errorContains: "can't get labels",
		},
		{
			name: "successful sync with label operations",
			repo: "test-repo",
			labels: []Label{
				{Name: "bug", Color: "red", Description: "Bug label"},
			},
			mockSetup: func(mockClient *mock_graphql.MockClient) {
				// Mock GetRepositoryIssueLabels call - returns empty labels so new label will be created
				first := mockClient.EXPECT().MakeRequest(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, req *graphql.Request, resp *graphql.Response) error {
						if data, ok := resp.Data.(*GetRepositoryIssueLabelsResponse); ok {
							data.Repository = GetRepositoryIssueLabelsRepository{
								Id: "repo123",
								Labels: GetRepositoryIssueLabelsRepositoryLabelsLabelConnection{
									Edges: []GetRepositoryIssueLabelsRepositoryLabelsLabelConnectionEdgesLabelEdge{},
									PageInfo: GetRepositoryIssueLabelsRepositoryLabelsLabelConnectionPageInfo{
										HasNextPage: false,
										EndCursor:   "",
										StartCursor: "",
									},
								},
							}
						}
						return nil
					},
				).Times(1)
				// Mock CreateLabel call
				mockClient.EXPECT().MakeRequest(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, req *graphql.Request, resp *graphql.Response) error {
						if data, ok := resp.Data.(*CreateLabelResponse); ok {
							data.CreateLabel = CreateLabelCreateLabelCreateLabelPayload{
								Label: CreateLabelCreateLabelCreateLabelPayloadLabel{
									Name:        "bug",
									Color:       "red",
									Description: "Bug label",
									CreatedAt:   time.Now(),
									UpdatedAt:   time.Now(),
									IsDefault:   false,
									Url:         url.URL{},
								},
							}
						}
						return nil
					},
				).Times(1).After(first)
			},
			expectError: false,
		},
		{
			name: "error during label update",
			repo: "test-repo",
			labels: []Label{
				{Name: "bug", Color: "red", Description: "Bug label"},
			},
			mockSetup: func(mockClient *mock_graphql.MockClient) {
				// Mock GetRepositoryIssueLabels call - returns existing label with same name
				first := mockClient.EXPECT().MakeRequest(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, req *graphql.Request, resp *graphql.Response) error {
						if data, ok := resp.Data.(*GetRepositoryIssueLabelsResponse); ok {
							data.Repository = GetRepositoryIssueLabelsRepository{
								Id: "repo123",
								Labels: GetRepositoryIssueLabelsRepositoryLabelsLabelConnection{
									Edges: []GetRepositoryIssueLabelsRepositoryLabelsLabelConnectionEdgesLabelEdge{
										{
											Node: GetRepositoryIssueLabelsRepositoryLabelsLabelConnectionEdgesLabelEdgeNodeLabel{
												Name:        "bug",
												Color:       "blue",
												Description: "Old bug label",
												Id:          "label123",
												CreatedAt:   time.Now(),
												UpdatedAt:   time.Now(),
												IsDefault:   false,
												Url:         url.URL{},
											},
										},
									},
									PageInfo: GetRepositoryIssueLabelsRepositoryLabelsLabelConnectionPageInfo{
										HasNextPage: false,
										EndCursor:   "",
										StartCursor: "",
									},
								},
							}
						}
						return nil
					},
				).Times(1)
				// Mock UpdateLabel call that fails
				mockClient.EXPECT().MakeRequest(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("update failed")).Times(1).After(first)
			},
			expectError:   true,
			errorContains: "can't update label",
		},
		{
			name: "error during label creation",
			repo: "test-repo",
			labels: []Label{
				{Name: "enhancement", Color: "blue", Description: "Enhancement label"},
			},
			mockSetup: func(mockClient *mock_graphql.MockClient) {
				// Mock GetRepositoryIssueLabels call - returns empty labels
				first := mockClient.EXPECT().MakeRequest(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, req *graphql.Request, resp *graphql.Response) error {
						if data, ok := resp.Data.(*GetRepositoryIssueLabelsResponse); ok {
							data.Repository = GetRepositoryIssueLabelsRepository{
								Id: "repo123",
								Labels: GetRepositoryIssueLabelsRepositoryLabelsLabelConnection{
									Edges: []GetRepositoryIssueLabelsRepositoryLabelsLabelConnectionEdgesLabelEdge{},
									PageInfo: GetRepositoryIssueLabelsRepositoryLabelsLabelConnectionPageInfo{
										HasNextPage: false,
										EndCursor:   "",
										StartCursor: "",
									},
								},
							}
						}
						return nil
					},
				).Times(1)
				// Mock CreateLabel call that fails
				mockClient.EXPECT().MakeRequest(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("create failed")).Times(1).After(first)
			},
			expectError:   true,
			errorContains: "can't create label",
		},
		{
			name:   "error during label deletion",
			repo:   "test-repo",
			labels: []Label{},
			mockSetup: func(mockClient *mock_graphql.MockClient) {
				// Mock GetRepositoryIssueLabels call - returns existing label that should be deleted
				first := mockClient.EXPECT().MakeRequest(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, req *graphql.Request, resp *graphql.Response) error {
						if data, ok := resp.Data.(*GetRepositoryIssueLabelsResponse); ok {
							data.Repository = GetRepositoryIssueLabelsRepository{
								Id: "repo123",
								Labels: GetRepositoryIssueLabelsRepositoryLabelsLabelConnection{
									Edges: []GetRepositoryIssueLabelsRepositoryLabelsLabelConnectionEdgesLabelEdge{
										{
											Node: GetRepositoryIssueLabelsRepositoryLabelsLabelConnectionEdgesLabelEdgeNodeLabel{
												Name:        "old-label",
												Color:       "gray",
												Description: "Old label to be deleted",
												Id:          "label456",
												CreatedAt:   time.Now(),
												UpdatedAt:   time.Now(),
												IsDefault:   false,
												Url:         url.URL{},
											},
										},
									},
									PageInfo: GetRepositoryIssueLabelsRepositoryLabelsLabelConnectionPageInfo{
										HasNextPage: false,
										EndCursor:   "",
										StartCursor: "",
									},
								},
							}
						}
						return nil
					},
				).Times(1)
				// Mock DeleteLabel call that fails
				mockClient.EXPECT().MakeRequest(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("delete failed")).Times(1).After(first)
			},
			expectError:   true,
			errorContains: "can't delete label",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := mock_graphql.NewMockClient(ctrl)
			ctx := context.Background()

			client := &GitHubV4Client{
				Context: ctx,
				client:  mockClient,
			}

			tt.mockSetup(mockClient)

			err := client.SyncLabels(tt.repo, "test-org", tt.labels)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.errorContains != "" && !containsString(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorContains, err)
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestGitHubV4Client_ContextCancellation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		contextFunc func() context.Context
		method      string
	}{
		{
			name: "canceled context for GetRateLimit",
			contextFunc: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			method: "GetRateLimit",
		},
		{
			name: "canceled context for GetTeams",
			contextFunc: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			method: "GetTeams",
		},
		{
			name: "canceled context for SyncLabels",
			contextFunc: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			method: "SyncLabels",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := mock_graphql.NewMockClient(ctrl)
			ctx := tt.contextFunc()

			client := &GitHubV4Client{
				Context: ctx,
				client:  mockClient,
			}

			// Mock the GraphQL call - context cancellation should be handled by the underlying client
			mockClient.EXPECT().MakeRequest(ctx, gomock.Any(), gomock.Any()).Return(context.Canceled).AnyTimes()

			switch tt.method {
			case getRateLimitMethod:
				_, err := client.GetRateLimit()
				if err == nil {
					t.Error("Expected context cancellation error")
				}
			case "GetTeams":
				_, err := client.GetTeams("test-org")
				if err == nil {
					t.Error("Expected context cancellation error")
				}
			case syncLabelsMethod:
				err := client.SyncLabels("test-repo", "test-org", []Label{})
				if err == nil {
					t.Error("Expected context cancellation error")
				}
			}
		})
	}
}

func TestGitHubV4Client_ErrorWrapping(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock_graphql.NewMockClient(ctrl)
	ctx := context.Background()

	client := &GitHubV4Client{
		Context: ctx,
		client:  mockClient,
	}

	tests := []struct {
		name        string
		method      string
		setupMock   func()
		expectError bool
		errorPrefix string
	}{
		{
			name:   "GetRateLimit error wrapping",
			method: "GetRateLimit",
			setupMock: func() {
				mockClient.EXPECT().MakeRequest(ctx, gomock.Any(), gomock.Any()).Return(errors.New("original error")).Times(1)
			},
			expectError: true,
			errorPrefix: "can't get ratelimit",
		},
		{
			name:   "SyncLabels error wrapping",
			method: "SyncLabels",
			setupMock: func() {
				mockClient.EXPECT().MakeRequest(ctx, gomock.Any(), gomock.Any()).Return(errors.New("original error")).Times(1)
			},
			expectError: true,
			errorPrefix: "can't get labels",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			var err error
			switch tt.method {
			case getRateLimitMethod:
				_, err = client.GetRateLimit()
			case syncLabelsMethod:
				err = client.SyncLabels("test-repo", "test-org", []Label{})
			}

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if !containsString(err.Error(), tt.errorPrefix) {
					t.Errorf("Expected error to start with '%s', got: %v", tt.errorPrefix, err)
				}
				// Test that the error is wrapped correctly
				if !containsString(err.Error(), "original error") {
					t.Errorf("Expected wrapped error to contain 'original error', got: %v", err)
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// Helper function to check if a string contains a substring.
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || substr == "" ||
		(len(s) > len(substr) && func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}()))
}
