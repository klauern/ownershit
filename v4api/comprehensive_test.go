package v4api

import (
	"context"
	"errors"
	"net/url"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	mock_graphql "github.com/klauern/ownershit/v4api/mocks"
)

// Comprehensive tests for generated types and business logic to reach 80%+ coverage

func TestGitHubV4Client_SyncLabels_CompleteScenarios(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock_graphql.NewMockClient(ctrl)
	ctx := context.Background()

	client := &GitHubV4Client{
		Context: ctx,
		client:  mockClient,
	}

	testURL, _ := url.Parse("https://github.com/test/labels/existing")
	testTime := time.Now()

	tests := []struct {
		name           string
		labels         []Label
		setupMocks     func()
		expectError    bool
		errorSubstring string
	}{
		{
			name: "successful sync with create, update, and delete operations",
			labels: []Label{
				{
					Name:        "new-label",
					Color:       "ff0000",
					Description: "New label to create",
				},
				{
					Name:        "existing-label",
					Color:       "00ff00",
					Description: "Updated description",
				},
			},
			setupMocks: func() {
				// Mock getting existing labels
				mockClient.EXPECT().MakeRequest(ctx, gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, req interface{}, resp interface{}) error {
						// Simulate response with existing labels
						if r, ok := resp.(*GetRepositoryIssueLabelsResponse); ok {
							r.Repository = GetRepositoryIssueLabelsRepository{
								Id: "repo-id-123",
								Labels: GetRepositoryIssueLabelsRepositoryLabelsLabelConnection{
									Edges: []GetRepositoryIssueLabelsRepositoryLabelsLabelConnectionEdgesLabelEdge{
										{
											Node: GetRepositoryIssueLabelsRepositoryLabelsLabelConnectionEdgesLabelEdgeNodeLabel{
												Name:        "existing-label",
												Color:       "ff0000", // Different color - will be updated
												Description: "Old description",
												Id:          "existing-label-id",
												CreatedAt:   testTime,
												UpdatedAt:   testTime,
												IsDefault:   false,
												Url:         *testURL,
											},
										},
										{
											Node: GetRepositoryIssueLabelsRepositoryLabelsLabelConnectionEdgesLabelEdgeNodeLabel{
												Name:        "label-to-delete",
												Color:       "0000ff",
												Description: "This will be deleted",
												Id:          "delete-label-id",
												CreatedAt:   testTime,
												UpdatedAt:   testTime,
												IsDefault:   false,
												Url:         *testURL,
											},
										},
									},
								},
							}
						}
						return nil
					},
				).Times(1)

				// Mock UpdateLabel call
				mockClient.EXPECT().MakeRequest(ctx, gomock.Any(), gomock.Any()).Return(nil).Times(1)

				// Mock CreateLabel call
				mockClient.EXPECT().MakeRequest(ctx, gomock.Any(), gomock.Any()).Return(nil).Times(1)

				// Mock DeleteLabel call
				mockClient.EXPECT().MakeRequest(ctx, gomock.Any(), gomock.Any()).Return(nil).Times(1)
			},
			expectError: false,
		},
		{
			name: "error getting existing labels",
			labels: []Label{
				{Name: "test-label", Color: "ff0000", Description: "Test"},
			},
			setupMocks: func() {
				mockClient.EXPECT().MakeRequest(ctx, gomock.Any(), gomock.Any()).Return(
					errors.New("failed to get repository labels"),
				).Times(1)
			},
			expectError:    true,
			errorSubstring: "can't get labels",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			err := client.SyncLabels("test-repo", "test-owner", tt.labels)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorSubstring != "" && !contains(err.Error(), tt.errorSubstring) {
					t.Errorf("Expected error to contain '%s', but got: %v", tt.errorSubstring, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestGitHubV4Client_GetTeams_Comprehensive(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock_graphql.NewMockClient(ctrl)
	ctx := context.Background()

	client := &GitHubV4Client{
		Context: ctx,
		client:  mockClient,
	}

	tests := []struct {
		name           string
		setupMocks     func()
		expectError    bool
		errorSubstring string
		expectedTeams  int
	}{
		{
			name: "successful multi-page pagination",
			setupMocks: func() {
				// First page
				mockClient.EXPECT().MakeRequest(ctx, gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, req interface{}, resp interface{}) error {
						if r, ok := resp.(*GetTeamsResponse); ok {
							r.Organization = GetTeamsOrganization{
								Teams: GetTeamsOrganizationTeamsTeamConnection{
									PageInfo: GetTeamsOrganizationTeamsTeamConnectionPageInfo{
										HasNextPage: true,
										EndCursor:   "cursor-1",
									},
									Edges: []GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdge{
										{
											Node: GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeam{
												Name:        "team1",
												Description: "First team",
												Members: GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamMembersTeamMemberConnection{
													Edges: []GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamMembersTeamMemberConnectionEdgesTeamMemberEdge{},
												},
												ChildTeams: GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamChildTeamsTeamConnection{
													Edges: []GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamChildTeamsTeamConnectionEdgesTeamEdge{},
												},
											},
										},
									},
								},
							}
						}
						return nil
					},
				).Times(1)

				// Second page
				mockClient.EXPECT().MakeRequest(ctx, gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, req interface{}, resp interface{}) error {
						if r, ok := resp.(*GetTeamsResponse); ok {
							r.Organization = GetTeamsOrganization{
								Teams: GetTeamsOrganizationTeamsTeamConnection{
									PageInfo: GetTeamsOrganizationTeamsTeamConnectionPageInfo{
										HasNextPage: false,
										EndCursor:   "cursor-2",
									},
									Edges: []GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdge{
										{
											Node: GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeam{
												Name:        "team2",
												Description: "Second team",
												Members: GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamMembersTeamMemberConnection{
													Edges: []GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamMembersTeamMemberConnectionEdgesTeamMemberEdge{},
												},
												ChildTeams: GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamChildTeamsTeamConnection{
													Edges: []GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamChildTeamsTeamConnectionEdgesTeamEdge{},
												},
											},
										},
									},
								},
							}
						}
						return nil
					},
				).Times(1)
			},
			expectError:   false,
			expectedTeams: 2,
		},
		{
			name: "error on initial request",
			setupMocks: func() {
				mockClient.EXPECT().MakeRequest(ctx, gomock.Any(), gomock.Any()).Return(
					errors.New("initial request failed"),
				).Times(1)
			},
			expectError:    true,
			errorSubstring: "failed to get teams (initial request)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			teams, err := client.GetTeams("test-org")

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorSubstring != "" && !contains(err.Error(), tt.errorSubstring) {
					t.Errorf("Expected error to contain '%s', but got: %v", tt.errorSubstring, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(teams) != tt.expectedTeams {
					t.Errorf("Expected %d teams, got %d", tt.expectedTeams, len(teams))
				}
			}
		})
	}
}

func TestGitHubV4Client_GetRateLimit_Comprehensive(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock_graphql.NewMockClient(ctrl)
	ctx := context.Background()

	client := &GitHubV4Client{
		Context: ctx,
		client:  mockClient,
	}

	tests := []struct {
		name           string
		setupMocks     func()
		expectError    bool
		errorSubstring string
		expectLimit    int
	}{
		{
			name: "successful rate limit retrieval",
			setupMocks: func() {
				resetTime := time.Now().Add(time.Hour)
				mockClient.EXPECT().MakeRequest(ctx, gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, req interface{}, resp interface{}) error {
						if r, ok := resp.(*GetRateLimitResponse); ok {
							r.RateLimit = GetRateLimitRateLimit{
								Limit:     5000,
								Cost:      1,
								Remaining: 4999,
								ResetAt:   resetTime,
							}
							r.Viewer = GetRateLimitViewerUser{
								Login: "test-user",
							}
						}
						return nil
					},
				).Times(1)
			},
			expectError: false,
			expectLimit: 5000,
		},
		{
			name: "rate limit API error",
			setupMocks: func() {
				mockClient.EXPECT().MakeRequest(ctx, gomock.Any(), gomock.Any()).Return(
					errors.New("rate limit API unavailable"),
				).Times(1)
			},
			expectError:    true,
			errorSubstring: "can't get ratelimit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			rateLimit, err := client.GetRateLimit()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorSubstring != "" && !contains(err.Error(), tt.errorSubstring) {
					t.Errorf("Expected error to contain '%s', but got: %v", tt.errorSubstring, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if rateLimit.RateLimit.Limit != tt.expectLimit {
					t.Errorf("Expected limit %d, got %d", tt.expectLimit, rateLimit.RateLimit.Limit)
				}
			}
		})
	}
}

// Helper function to check if error message contains substring
func contains(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Tests for generated type getters to improve coverage significantly

func TestGetRepositoryIssueLabelsRepository_Getters(t *testing.T) {
	testTime := time.Now()
	testURL, _ := url.Parse("https://github.com/test/labels/repo")

	repo := &GetRepositoryIssueLabelsRepository{
		Id: "repo-id-123",
		Labels: GetRepositoryIssueLabelsRepositoryLabelsLabelConnection{
			Edges: []GetRepositoryIssueLabelsRepositoryLabelsLabelConnectionEdgesLabelEdge{
				{
					Node: GetRepositoryIssueLabelsRepositoryLabelsLabelConnectionEdgesLabelEdgeNodeLabel{
						Name:        "test-label",
						CreatedAt:   testTime,
						Color:       "ff0000",
						Description: "Test label description",
						IsDefault:   true,
						UpdatedAt:   testTime,
						Url:         *testURL,
						Id:          "label-id-456",
					},
				},
			},
			PageInfo: GetRepositoryIssueLabelsRepositoryLabelsLabelConnectionPageInfo{
				EndCursor:   "end-cursor-123",
				StartCursor: "start-cursor-123",
				HasNextPage: false,
			},
		},
	}

	if repo.GetId() != "repo-id-123" {
		t.Errorf("GetId() = %v, want %v", repo.GetId(), "repo-id-123")
	}

	labels := repo.GetLabels()
	if len(labels.Edges) != 1 {
		t.Errorf("GetLabels().Edges length = %v, want %v", len(labels.Edges), 1)
	}
}

func TestGetRepositoryIssueLabelsRepositoryLabelsLabelConnection_Getters(t *testing.T) {
	connection := &GetRepositoryIssueLabelsRepositoryLabelsLabelConnection{
		Edges: []GetRepositoryIssueLabelsRepositoryLabelsLabelConnectionEdgesLabelEdge{},
		PageInfo: GetRepositoryIssueLabelsRepositoryLabelsLabelConnectionPageInfo{
			EndCursor:   "connection-end",
			StartCursor: "connection-start",
			HasNextPage: true,
		},
	}

	edges := connection.GetEdges()
	if len(edges) != 0 {
		t.Errorf("GetEdges() length = %v, want %v", len(edges), 0)
	}

	pageInfo := connection.GetPageInfo()
	if pageInfo.EndCursor != "connection-end" {
		t.Errorf("GetPageInfo().EndCursor = %v, want %v", pageInfo.EndCursor, "connection-end")
	}
}

func TestGetRepositoryIssueLabelsRepositoryLabelsLabelConnectionPageInfo_Getters(t *testing.T) {
	pageInfo := &GetRepositoryIssueLabelsRepositoryLabelsLabelConnectionPageInfo{
		EndCursor:   "page-end-cursor",
		StartCursor: "page-start-cursor",
		HasNextPage: true,
	}

	if pageInfo.GetEndCursor() != "page-end-cursor" {
		t.Errorf("GetEndCursor() = %v, want %v", pageInfo.GetEndCursor(), "page-end-cursor")
	}
	if pageInfo.GetStartCursor() != "page-start-cursor" {
		t.Errorf("GetStartCursor() = %v, want %v", pageInfo.GetStartCursor(), "page-start-cursor")
	}
	if !pageInfo.GetHasNextPage() {
		t.Errorf("GetHasNextPage() = %v, want %v", pageInfo.GetHasNextPage(), true)
	}
}

func TestGetTeamsOrganization_GetTeams(t *testing.T) {
	org := &GetTeamsOrganization{
		Teams: GetTeamsOrganizationTeamsTeamConnection{
			PageInfo: GetTeamsOrganizationTeamsTeamConnectionPageInfo{
				HasNextPage: false,
				StartCursor: "org-start",
				EndCursor:   "org-end",
			},
			Edges: []GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdge{},
		},
	}

	teams := org.GetTeams()
	if len(teams.Edges) != 0 {
		t.Errorf("GetTeams().Edges length = %v, want %v", len(teams.Edges), 0)
	}
}

func TestGetTeamsOrganizationTeamsTeamConnection_Getters(t *testing.T) {
	connection := &GetTeamsOrganizationTeamsTeamConnection{
		PageInfo: GetTeamsOrganizationTeamsTeamConnectionPageInfo{
			HasNextPage: true,
			StartCursor: "team-connection-start",
			EndCursor:   "team-connection-end",
		},
		Edges: []GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdge{},
	}

	pageInfo := connection.GetPageInfo()
	if pageInfo.HasNextPage != true {
		t.Errorf("GetPageInfo().HasNextPage = %v, want %v", pageInfo.HasNextPage, true)
	}

	edges := connection.GetEdges()
	if len(edges) != 0 {
		t.Errorf("GetEdges() length = %v, want %v", len(edges), 0)
	}
}

func TestGetTeamsOrganizationTeamsTeamConnectionPageInfo_Getters(t *testing.T) {
	pageInfo := &GetTeamsOrganizationTeamsTeamConnectionPageInfo{
		HasNextPage: true,
		StartCursor: "teams-page-start",
		EndCursor:   "teams-page-end",
	}

	if !pageInfo.GetHasNextPage() {
		t.Errorf("GetHasNextPage() = %v, want %v", pageInfo.GetHasNextPage(), true)
	}
	if pageInfo.GetStartCursor() != "teams-page-start" {
		t.Errorf("GetStartCursor() = %v, want %v", pageInfo.GetStartCursor(), "teams-page-start")
	}
	if pageInfo.GetEndCursor() != "teams-page-end" {
		t.Errorf("GetEndCursor() = %v, want %v", pageInfo.GetEndCursor(), "teams-page-end")
	}
}

func TestGetTeamsResponse_Getters(t *testing.T) {
	resetTime := time.Now().Add(time.Hour)
	response := &GetTeamsResponse{
		Organization: GetTeamsOrganization{
			Teams: GetTeamsOrganizationTeamsTeamConnection{
				PageInfo: GetTeamsOrganizationTeamsTeamConnectionPageInfo{
					HasNextPage: false,
					EndCursor:   "response-cursor",
				},
				Edges: []GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdge{},
			},
		},
		Viewer: GetTeamsViewerUser{
			Login: "response-viewer",
		},
		RateLimit: GetTeamsRateLimit{
			Limit:     5000,
			Cost:      1,
			Remaining: 4998,
			ResetAt:   resetTime,
		},
	}

	org := response.GetOrganization()
	if len(org.Teams.Edges) != 0 {
		t.Errorf("GetOrganization().Teams.Edges length = %v, want %v", len(org.Teams.Edges), 0)
	}

	viewer := response.GetViewer()
	if viewer.Login != "response-viewer" {
		t.Errorf("GetViewer().Login = %v, want %v", viewer.Login, "response-viewer")
	}

	rateLimit := response.GetRateLimit()
	if rateLimit.Limit != 5000 {
		t.Errorf("GetRateLimit().Limit = %v, want %v", rateLimit.Limit, 5000)
	}
}

func TestGetTeamsViewerUser_GetLogin(t *testing.T) {
	viewer := &GetTeamsViewerUser{
		Login: "viewer-user-login",
	}

	if viewer.GetLogin() != "viewer-user-login" {
		t.Errorf("GetLogin() = %v, want %v", viewer.GetLogin(), "viewer-user-login")
	}
}

func TestGetTeamsRateLimit_Getters(t *testing.T) {
	resetTime := time.Now().Add(2 * time.Hour)
	rateLimit := &GetTeamsRateLimit{
		Limit:     5000,
		Cost:      2,
		Remaining: 4997,
		ResetAt:   resetTime,
	}

	if rateLimit.GetLimit() != 5000 {
		t.Errorf("GetLimit() = %v, want %v", rateLimit.GetLimit(), 5000)
	}
	if rateLimit.GetCost() != 2 {
		t.Errorf("GetCost() = %v, want %v", rateLimit.GetCost(), 2)
	}
	if rateLimit.GetRemaining() != 4997 {
		t.Errorf("GetRemaining() = %v, want %v", rateLimit.GetRemaining(), 4997)
	}
	if !rateLimit.GetResetAt().Equal(resetTime) {
		t.Errorf("GetResetAt() = %v, want %v", rateLimit.GetResetAt(), resetTime)
	}
}

func TestTeamOrder_Getters(t *testing.T) {
	order := &TeamOrder{
		Direction: OrderDirectionAsc,
		Field:     TeamOrderFieldName,
	}

	if order.GetDirection() != OrderDirectionAsc {
		t.Errorf("GetDirection() = %v, want %v", order.GetDirection(), OrderDirectionAsc)
	}
	if order.GetField() != TeamOrderFieldName {
		t.Errorf("GetField() = %v, want %v", order.GetField(), TeamOrderFieldName)
	}
}

func TestUpdateLabelInput_Getters(t *testing.T) {
	input := &UpdateLabelInput{
		ClientMutationId: "update-456",
		Color:            "purple",
		Description:      "Updated purple label",
		Id:               "update-id",
		Name:             "updated-name",
	}

	if input.GetClientMutationId() != "update-456" {
		t.Errorf("GetClientMutationId() = %v, want %v", input.GetClientMutationId(), "update-456")
	}
	if input.GetColor() != "purple" {
		t.Errorf("GetColor() = %v, want %v", input.GetColor(), "purple")
	}
	if input.GetDescription() != "Updated purple label" {
		t.Errorf("GetDescription() = %v, want %v", input.GetDescription(), "Updated purple label")
	}
	if input.GetId() != "update-id" {
		t.Errorf("GetId() = %v, want %v", input.GetId(), "update-id")
	}
	if input.GetName() != "updated-name" {
		t.Errorf("GetName() = %v, want %v", input.GetName(), "updated-name")
	}
}

func TestUpdateLabelResponse_GetUpdateLabel(t *testing.T) {
	testTime := time.Now()
	testURL, _ := url.Parse("https://github.com/test/updated")

	response := &UpdateLabelResponse{
		UpdateLabel: UpdateLabelUpdateLabelUpdateLabelPayload{
			Label: UpdateLabelUpdateLabelUpdateLabelPayloadLabel{
				Name:        "updated-label",
				CreatedAt:   testTime,
				Color:       "cyan",
				Description: "Updated label description",
				IsDefault:   true,
				UpdatedAt:   testTime,
				Url:         *testURL,
			},
		},
	}

	result := response.GetUpdateLabel()
	if result.Label.Name != "updated-label" {
		t.Errorf("GetUpdateLabel().Label.Name = %v, want %v", result.Label.Name, "updated-label")
	}
}

func TestUpdateLabelUpdateLabelUpdateLabelPayload_GetLabel(t *testing.T) {
	testTime := time.Now()
	testURL, _ := url.Parse("https://github.com/test/payload")

	payload := &UpdateLabelUpdateLabelUpdateLabelPayload{
		Label: UpdateLabelUpdateLabelUpdateLabelPayloadLabel{
			Name:        "payload-label",
			CreatedAt:   testTime,
			Color:       "orange",
			Description: "Payload description",
			IsDefault:   false,
			UpdatedAt:   testTime,
			Url:         *testURL,
		},
	}

	label := payload.GetLabel()
	if label.Name != "payload-label" {
		t.Errorf("GetLabel().Name = %v, want %v", label.Name, "payload-label")
	}
}

func TestUpdateLabelUpdateLabelUpdateLabelPayloadLabel_Getters(t *testing.T) {
	testTime := time.Now()
	testURL, _ := url.Parse("https://github.com/test/comprehensive")

	label := &UpdateLabelUpdateLabelUpdateLabelPayloadLabel{
		Name:        "comprehensive-update",
		CreatedAt:   testTime,
		Color:       "magenta",
		Description: "Comprehensive update test",
		IsDefault:   true,
		UpdatedAt:   testTime,
		Url:         *testURL,
	}

	if label.GetName() != "comprehensive-update" {
		t.Errorf("GetName() = %v, want %v", label.GetName(), "comprehensive-update")
	}
	if !label.GetCreatedAt().Equal(testTime) {
		t.Errorf("GetCreatedAt() = %v, want %v", label.GetCreatedAt(), testTime)
	}
	if label.GetColor() != "magenta" {
		t.Errorf("GetColor() = %v, want %v", label.GetColor(), "magenta")
	}
	if label.GetDescription() != "Comprehensive update test" {
		t.Errorf("GetDescription() = %v, want %v", label.GetDescription(), "Comprehensive update test")
	}
	if !label.GetIsDefault() {
		t.Errorf("GetIsDefault() = %v, want %v", label.GetIsDefault(), true)
	}
	if !label.GetUpdatedAt().Equal(testTime) {
		t.Errorf("GetUpdatedAt() = %v, want %v", label.GetUpdatedAt(), testTime)
	}
	if label.GetUrl() != *testURL {
		t.Errorf("GetUrl() = %v, want %v", label.GetUrl(), *testURL)
	}
}