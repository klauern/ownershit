package v4api

import (
	"net/url"
	"testing"
	"time"
)

// Additional tests to boost coverage to 80%+

// Test all remaining getter methods from generated.go

func TestCreateLabelCreateLabelCreateLabelPayload_GetLabel(t *testing.T) {
	testTime := time.Now()
	testURL, _ := url.Parse("https://github.com/test/label")

	payload := &CreateLabelCreateLabelCreateLabelPayload{
		Label: CreateLabelCreateLabelCreateLabelPayloadLabel{
			Name:        "test-label",
			CreatedAt:   testTime,
			Color:       "ff0000",
			Description: "Test label",
			IsDefault:   true,
			UpdatedAt:   testTime,
			Url:         *testURL,
		},
	}

	result := payload.GetLabel()
	if result.Name != "test-label" {
		t.Errorf("GetLabel().Name = %v, want %v", result.Name, "test-label")
	}
}

func TestCreateLabelCreateLabelCreateLabelPayloadLabel_Getters(t *testing.T) {
	testTime := time.Now()
	testURL, _ := url.Parse("https://github.com/test/label")

	label := &CreateLabelCreateLabelCreateLabelPayloadLabel{
		Name:        "test-label",
		CreatedAt:   testTime,
		Color:       "ff0000",
		Description: "Test label description",
		IsDefault:   true,
		UpdatedAt:   testTime,
		Url:         *testURL,
	}

	if label.GetName() != "test-label" {
		t.Errorf("GetName() = %v, want %v", label.GetName(), "test-label")
	}
	if !label.GetCreatedAt().Equal(testTime) {
		t.Errorf("GetCreatedAt() = %v, want %v", label.GetCreatedAt(), testTime)
	}
	if label.GetColor() != "ff0000" {
		t.Errorf("GetColor() = %v, want %v", label.GetColor(), "ff0000")
	}
	if label.GetDescription() != "Test label description" {
		t.Errorf("GetDescription() = %v, want %v", label.GetDescription(), "Test label description")
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

func TestCreateLabelInput_Getters(t *testing.T) {
	input := &CreateLabelInput{
		ClientMutationId: "mutation-123",
		Color:            "00ff00",
		Description:      "Green label",
		Name:             "green-label",
		RepositoryId:     "repo-123",
	}

	if input.GetClientMutationId() != "mutation-123" {
		t.Errorf("GetClientMutationId() = %v, want %v", input.GetClientMutationId(), "mutation-123")
	}
	if input.GetColor() != "00ff00" {
		t.Errorf("GetColor() = %v, want %v", input.GetColor(), "00ff00")
	}
	if input.GetDescription() != "Green label" {
		t.Errorf("GetDescription() = %v, want %v", input.GetDescription(), "Green label")
	}
	if input.GetName() != "green-label" {
		t.Errorf("GetName() = %v, want %v", input.GetName(), "green-label")
	}
	if input.GetRepositoryId() != "repo-123" {
		t.Errorf("GetRepositoryId() = %v, want %v", input.GetRepositoryId(), "repo-123")
	}
}

func TestCreateLabelResponse_GetCreateLabel(t *testing.T) {
	testTime := time.Now()
	testURL, _ := url.Parse("https://github.com/test/label")

	response := &CreateLabelResponse{
		CreateLabel: CreateLabelCreateLabelCreateLabelPayload{
			Label: CreateLabelCreateLabelCreateLabelPayloadLabel{
				Name:        "response-label",
				CreatedAt:   testTime,
				Color:       "0000ff",
				Description: "Response label",
				IsDefault:   false,
				UpdatedAt:   testTime,
				Url:         *testURL,
			},
		},
	}

	result := response.GetCreateLabel()
	if result.Label.Name != "response-label" {
		t.Errorf("GetCreateLabel().Label.Name = %v, want %v", result.Label.Name, "response-label")
	}
}

func TestDeleteLabelDeleteLabelDeleteLabelPayload_GetClientMutationId(t *testing.T) {
	payload := &DeleteLabelDeleteLabelDeleteLabelPayload{
		ClientMutationId: "delete-123",
	}

	if payload.GetClientMutationId() != "delete-123" {
		t.Errorf("GetClientMutationId() = %v, want %v", payload.GetClientMutationId(), "delete-123")
	}
}

func TestDeleteLabelInput_Getters(t *testing.T) {
	input := &DeleteLabelInput{
		ClientMutationId: "delete-mutation-123",
		Id:               "label-id-456",
	}

	if input.GetClientMutationId() != "delete-mutation-123" {
		t.Errorf("GetClientMutationId() = %v, want %v", input.GetClientMutationId(), "delete-mutation-123")
	}
	if input.GetId() != "label-id-456" {
		t.Errorf("GetId() = %v, want %v", input.GetId(), "label-id-456")
	}
}

func TestDeleteLabelResponse_GetDeleteLabel(t *testing.T) {
	response := &DeleteLabelResponse{
		DeleteLabel: DeleteLabelDeleteLabelDeleteLabelPayload{
			ClientMutationId: "delete-response-123",
		},
	}

	result := response.GetDeleteLabel()
	if result.ClientMutationId != "delete-response-123" {
		t.Errorf("GetDeleteLabel().ClientMutationId = %v, want %v", result.ClientMutationId, "delete-response-123")
	}
}

func TestGetRateLimitRateLimit_Getters(t *testing.T) {
	resetTime := time.Now().Add(time.Hour)
	rateLimit := &GetRateLimitRateLimit{
		Limit:     5000,
		Cost:      1,
		Remaining: 4999,
		ResetAt:   resetTime,
	}

	if rateLimit.GetLimit() != 5000 {
		t.Errorf("GetLimit() = %v, want %v", rateLimit.GetLimit(), 5000)
	}
	if rateLimit.GetCost() != 1 {
		t.Errorf("GetCost() = %v, want %v", rateLimit.GetCost(), 1)
	}
	if rateLimit.GetRemaining() != 4999 {
		t.Errorf("GetRemaining() = %v, want %v", rateLimit.GetRemaining(), 4999)
	}
	if !rateLimit.GetResetAt().Equal(resetTime) {
		t.Errorf("GetResetAt() = %v, want %v", rateLimit.GetResetAt(), resetTime)
	}
}

func TestGetRateLimitResponse_Getters(t *testing.T) {
	resetTime := time.Now().Add(time.Hour)
	response := &GetRateLimitResponse{
		Viewer: GetRateLimitViewerUser{
			Login: "test-user",
		},
		RateLimit: GetRateLimitRateLimit{
			Limit:     5000,
			Cost:      1,
			Remaining: 4999,
			ResetAt:   resetTime,
		},
	}

	viewer := response.GetViewer()
	if viewer.Login != "test-user" {
		t.Errorf("GetViewer().Login = %v, want %v", viewer.Login, "test-user")
	}

	rateLimit := response.GetRateLimit()
	if rateLimit.Limit != 5000 {
		t.Errorf("GetRateLimit().Limit = %v, want %v", rateLimit.Limit, 5000)
	}
}

func TestGetRateLimitViewerUser_GetLogin(t *testing.T) {
	viewer := &GetRateLimitViewerUser{
		Login: "viewer-login",
	}

	if viewer.GetLogin() != "viewer-login" {
		t.Errorf("GetLogin() = %v, want %v", viewer.GetLogin(), "viewer-login")
	}
}

func TestGetRepositoryIssueLabelsRepositoryLabelsLabelConnectionEdgesLabelEdge_GetNode(t *testing.T) {
	testTime := time.Now()
	testURL, _ := url.Parse("https://github.com/test/label")

	edge := &GetRepositoryIssueLabelsRepositoryLabelsLabelConnectionEdgesLabelEdge{
		Node: GetRepositoryIssueLabelsRepositoryLabelsLabelConnectionEdgesLabelEdgeNodeLabel{
			Name:        "edge-label",
			CreatedAt:   testTime,
			Color:       "purple",
			Description: "Edge label description",
			IsDefault:   false,
			UpdatedAt:   testTime,
			Url:         *testURL,
			Id:          "edge-label-id",
		},
	}

	node := edge.GetNode()
	if node.Name != "edge-label" {
		t.Errorf("GetNode().Name = %v, want %v", node.Name, "edge-label")
	}
}

func TestGetRepositoryIssueLabelsRepositoryLabelsLabelConnectionEdgesLabelEdgeNodeLabel_Getters(t *testing.T) {
	testTime := time.Now()
	testURL, _ := url.Parse("https://github.com/test/label")

	label := &GetRepositoryIssueLabelsRepositoryLabelsLabelConnectionEdgesLabelEdgeNodeLabel{
		Name:        "node-label",
		CreatedAt:   testTime,
		Color:       "yellow",
		Description: "Node label description",
		IsDefault:   true,
		UpdatedAt:   testTime,
		Url:         *testURL,
		Id:          "node-label-id",
	}

	if label.GetName() != "node-label" {
		t.Errorf("GetName() = %v, want %v", label.GetName(), "node-label")
	}
	if !label.GetCreatedAt().Equal(testTime) {
		t.Errorf("GetCreatedAt() = %v, want %v", label.GetCreatedAt(), testTime)
	}
	if label.GetColor() != "yellow" {
		t.Errorf("GetColor() = %v, want %v", label.GetColor(), "yellow")
	}
	if label.GetDescription() != "Node label description" {
		t.Errorf("GetDescription() = %v, want %v", label.GetDescription(), "Node label description")
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
	if label.GetId() != "node-label-id" {
		t.Errorf("GetId() = %v, want %v", label.GetId(), "node-label-id")
	}
}

func TestGetRepositoryIssueLabelsResponse_GetRepository(t *testing.T) {
	response := &GetRepositoryIssueLabelsResponse{
		Repository: GetRepositoryIssueLabelsRepository{
			Id: "response-repo-id",
			Labels: GetRepositoryIssueLabelsRepositoryLabelsLabelConnection{
				Edges: []GetRepositoryIssueLabelsRepositoryLabelsLabelConnectionEdgesLabelEdge{},
			},
		},
	}

	repo := response.GetRepository()
	if repo.Id != "response-repo-id" {
		t.Errorf("GetRepository().Id = %v, want %v", repo.Id, "response-repo-id")
	}
}

func TestGetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdge_GetNode(t *testing.T) {
	edge := &GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdge{
		Node: GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeam{
			Name: "test-team",
			Members: GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamMembersTeamMemberConnection{
				Edges: []GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamMembersTeamMemberConnectionEdgesTeamMemberEdge{},
			},
			ChildTeams: GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamChildTeamsTeamConnection{
				Edges: []GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamChildTeamsTeamConnectionEdgesTeamEdge{},
			},
			Description: "Test team description",
		},
	}

	node := edge.GetNode()
	if node.Name != "test-team" {
		t.Errorf("GetNode().Name = %v, want %v", node.Name, "test-team")
	}
}

func TestGetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeam_Getters(t *testing.T) {
	team := &GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeam{
		Name: "development-team",
		Members: GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamMembersTeamMemberConnection{
			Edges: []GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamMembersTeamMemberConnectionEdgesTeamMemberEdge{},
		},
		ChildTeams: GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamChildTeamsTeamConnection{
			Edges: []GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamChildTeamsTeamConnectionEdgesTeamEdge{},
		},
		Description: "Development team description",
	}

	if team.GetName() != "development-team" {
		t.Errorf("GetName() = %v, want %v", team.GetName(), "development-team")
	}

	members := team.GetMembers()
	if len(members.Edges) != 0 {
		t.Errorf("GetMembers().Edges length = %v, want %v", len(members.Edges), 0)
	}

	childTeams := team.GetChildTeams()
	if len(childTeams.Edges) != 0 {
		t.Errorf("GetChildTeams().Edges length = %v, want %v", len(childTeams.Edges), 0)
	}

	if team.GetDescription() != "Development team description" {
		t.Errorf("GetDescription() = %v, want %v", team.GetDescription(), "Development team description")
	}
}

func TestGetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamMembersTeamMemberConnection_GetEdges(t *testing.T) {
	connection := &GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamMembersTeamMemberConnection{
		Edges: []GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamMembersTeamMemberConnectionEdgesTeamMemberEdge{},
	}

	edges := connection.GetEdges()
	if len(edges) != 0 {
		t.Errorf("GetEdges() length = %v, want %v", len(edges), 0)
	}
}

func TestGetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamMembersTeamMemberConnectionEdgesTeamMemberEdge_GetNode(t *testing.T) {
	edge := &GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamMembersTeamMemberConnectionEdgesTeamMemberEdge{
		Node: GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamMembersTeamMemberConnectionEdgesTeamMemberEdgeNodeUser{
			Id: "user-id-123",
		},
	}

	node := edge.GetNode()
	if node.Id != "user-id-123" {
		t.Errorf("GetNode().Id = %v, want %v", node.Id, "user-id-123")
	}
}

func TestGetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamMembersTeamMemberConnectionEdgesTeamMemberEdgeNodeUser_GetId(t *testing.T) {
	user := &GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamMembersTeamMemberConnectionEdgesTeamMemberEdgeNodeUser{
		Id: "user-456",
	}

	if user.GetId() != "user-456" {
		t.Errorf("GetId() = %v, want %v", user.GetId(), "user-456")
	}
}

func TestGetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamChildTeamsTeamConnection_GetEdges(t *testing.T) {
	connection := &GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamChildTeamsTeamConnection{
		Edges: []GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamChildTeamsTeamConnectionEdgesTeamEdge{},
	}

	edges := connection.GetEdges()
	if len(edges) != 0 {
		t.Errorf("GetEdges() length = %v, want %v", len(edges), 0)
	}
}

func TestGetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamChildTeamsTeamConnectionEdgesTeamEdge_GetNode(t *testing.T) {
	edge := &GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamChildTeamsTeamConnectionEdgesTeamEdge{
		Node: GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamChildTeamsTeamConnectionEdgesTeamEdgeNodeTeam{
			Id: "child-team-id-789",
		},
	}

	node := edge.GetNode()
	if node.Id != "child-team-id-789" {
		t.Errorf("GetNode().Id = %v, want %v", node.Id, "child-team-id-789")
	}
}

func TestGetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamChildTeamsTeamConnectionEdgesTeamEdgeNodeTeam_GetId(t *testing.T) {
	team := &GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdgeNodeTeamChildTeamsTeamConnectionEdgesTeamEdgeNodeTeam{
		Id: "team-987",
	}

	if team.GetId() != "team-987" {
		t.Errorf("GetId() = %v, want %v", team.GetId(), "team-987")
	}
}

// Test input variable getters that haven't been covered

func TestInputVariableGetters(t *testing.T) {
	createInput := __CreateLabelInput{
		Input: CreateLabelInput{
			Name:         "test-input-label",
			Color:        "red",
			Description:  "Input test label",
			RepositoryId: "input-repo-id",
		},
	}

	if createInput.GetInput().Name != "test-input-label" {
		t.Errorf("__CreateLabelInput.GetInput().Name = %v, want %v", createInput.GetInput().Name, "test-input-label")
	}

	deleteInput := __DeleteLabelInput{
		Input: DeleteLabelInput{
			Id: "delete-input-id",
		},
	}

	if deleteInput.GetInput().Id != "delete-input-id" {
		t.Errorf("__DeleteLabelInput.GetInput().Id = %v, want %v", deleteInput.GetInput().Id, "delete-input-id")
	}

	updateInput := __UpdateLabelInput{
		Input: UpdateLabelInput{
			Id:   "update-input-id",
			Name: "updated-input-label",
		},
	}

	if updateInput.GetInput().Name != "updated-input-label" {
		t.Errorf("__UpdateLabelInput.GetInput().Name = %v, want %v", updateInput.GetInput().Name, "updated-input-label")
	}

	repoInput := __GetRepositoryIssueLabelsInput{
		Name:   "input-repo",
		Owner:  "input-owner",
		Cursor: "input-cursor",
	}

	if repoInput.GetName() != "input-repo" {
		t.Errorf("__GetRepositoryIssueLabelsInput.GetName() = %v, want %v", repoInput.GetName(), "input-repo")
	}
	if repoInput.GetOwner() != "input-owner" {
		t.Errorf("__GetRepositoryIssueLabelsInput.GetOwner() = %v, want %v", repoInput.GetOwner(), "input-owner")
	}
	if repoInput.GetCursor() != "input-cursor" {
		t.Errorf("__GetRepositoryIssueLabelsInput.GetCursor() = %v, want %v", repoInput.GetCursor(), "input-cursor")
	}

	teamsInput := __GetTeamsInput{
		Order:        TeamOrder{Direction: OrderDirectionAsc, Field: TeamOrderFieldName},
		First:        50,
		Cursor:       "teams-cursor",
		Organization: "teams-org",
	}

	if teamsInput.GetOrder().Direction != OrderDirectionAsc {
		t.Errorf("__GetTeamsInput.GetOrder().Direction = %v, want %v", teamsInput.GetOrder().Direction, OrderDirectionAsc)
	}
	if teamsInput.GetFirst() != 50 {
		t.Errorf("__GetTeamsInput.GetFirst() = %v, want %v", teamsInput.GetFirst(), 50)
	}
	if teamsInput.GetCursor() != "teams-cursor" {
		t.Errorf("__GetTeamsInput.GetCursor() = %v, want %v", teamsInput.GetCursor(), "teams-cursor")
	}
	if teamsInput.GetOrganization() != "teams-org" {
		t.Errorf("__GetTeamsInput.GetOrganization() = %v, want %v", teamsInput.GetOrganization(), "teams-org")
	}
}