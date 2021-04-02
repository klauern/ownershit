// Code generated by MockGen. DO NOT EDIT.
// Source: github_v3.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	github "github.com/google/go-github/v34/github"
)

// MockTeamsService is a mock of TeamsService interface.
type MockTeamsService struct {
	ctrl     *gomock.Controller
	recorder *MockTeamsServiceMockRecorder
}

// MockTeamsServiceMockRecorder is the mock recorder for MockTeamsService.
type MockTeamsServiceMockRecorder struct {
	mock *MockTeamsService
}

// NewMockTeamsService creates a new mock instance.
func NewMockTeamsService(ctrl *gomock.Controller) *MockTeamsService {
	mock := &MockTeamsService{ctrl: ctrl}
	mock.recorder = &MockTeamsServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTeamsService) EXPECT() *MockTeamsServiceMockRecorder {
	return m.recorder
}

// AddTeamRepoBySlug mocks base method.
func (m *MockTeamsService) AddTeamRepoBySlug(ctx context.Context, org, slug, owner, repo string, opts *github.TeamAddTeamRepoOptions) (*github.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddTeamRepoBySlug", ctx, org, slug, owner, repo, opts)
	ret0, _ := ret[0].(*github.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AddTeamRepoBySlug indicates an expected call of AddTeamRepoBySlug.
func (mr *MockTeamsServiceMockRecorder) AddTeamRepoBySlug(ctx, org, slug, owner, repo, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddTeamRepoBySlug", reflect.TypeOf((*MockTeamsService)(nil).AddTeamRepoBySlug), ctx, org, slug, owner, repo, opts)
}

// GetTeamBySlug mocks base method.
func (m *MockTeamsService) GetTeamBySlug(ctx context.Context, org, slug string) (*github.Team, *github.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTeamBySlug", ctx, org, slug)
	ret0, _ := ret[0].(*github.Team)
	ret1, _ := ret[1].(*github.Response)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetTeamBySlug indicates an expected call of GetTeamBySlug.
func (mr *MockTeamsServiceMockRecorder) GetTeamBySlug(ctx, org, slug interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTeamBySlug", reflect.TypeOf((*MockTeamsService)(nil).GetTeamBySlug), ctx, org, slug)
}

// MockIssuesService is a mock of IssuesService interface.
type MockIssuesService struct {
	ctrl     *gomock.Controller
	recorder *MockIssuesServiceMockRecorder
}

// MockIssuesServiceMockRecorder is the mock recorder for MockIssuesService.
type MockIssuesServiceMockRecorder struct {
	mock *MockIssuesService
}

// NewMockIssuesService creates a new mock instance.
func NewMockIssuesService(ctrl *gomock.Controller) *MockIssuesService {
	mock := &MockIssuesService{ctrl: ctrl}
	mock.recorder = &MockIssuesServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIssuesService) EXPECT() *MockIssuesServiceMockRecorder {
	return m.recorder
}

// CreateLabel mocks base method.
func (m *MockIssuesService) CreateLabel(ctx context.Context, owner, repo string, label *github.Label) (*github.Label, *github.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateLabel", ctx, owner, repo, label)
	ret0, _ := ret[0].(*github.Label)
	ret1, _ := ret[1].(*github.Response)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// CreateLabel indicates an expected call of CreateLabel.
func (mr *MockIssuesServiceMockRecorder) CreateLabel(ctx, owner, repo, label interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateLabel", reflect.TypeOf((*MockIssuesService)(nil).CreateLabel), ctx, owner, repo, label)
}

// EditLabel mocks base method.
func (m *MockIssuesService) EditLabel(ctx context.Context, owner, repo, name string, label *github.Label) (*github.Label, *github.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EditLabel", ctx, owner, repo, name, label)
	ret0, _ := ret[0].(*github.Label)
	ret1, _ := ret[1].(*github.Response)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// EditLabel indicates an expected call of EditLabel.
func (mr *MockIssuesServiceMockRecorder) EditLabel(ctx, owner, repo, name, label interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EditLabel", reflect.TypeOf((*MockIssuesService)(nil).EditLabel), ctx, owner, repo, name, label)
}

// ListLabels mocks base method.
func (m *MockIssuesService) ListLabels(ctx context.Context, owner, repo string, opts *github.ListOptions) ([]*github.Label, *github.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListLabels", ctx, owner, repo, opts)
	ret0, _ := ret[0].([]*github.Label)
	ret1, _ := ret[1].(*github.Response)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// ListLabels indicates an expected call of ListLabels.
func (mr *MockIssuesServiceMockRecorder) ListLabels(ctx, owner, repo, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListLabels", reflect.TypeOf((*MockIssuesService)(nil).ListLabels), ctx, owner, repo, opts)
}

// MockRepositoriesService is a mock of RepositoriesService interface.
type MockRepositoriesService struct {
	ctrl     *gomock.Controller
	recorder *MockRepositoriesServiceMockRecorder
}

// MockRepositoriesServiceMockRecorder is the mock recorder for MockRepositoriesService.
type MockRepositoriesServiceMockRecorder struct {
	mock *MockRepositoriesService
}

// NewMockRepositoriesService creates a new mock instance.
func NewMockRepositoriesService(ctrl *gomock.Controller) *MockRepositoriesService {
	mock := &MockRepositoriesService{ctrl: ctrl}
	mock.recorder = &MockRepositoriesServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRepositoriesService) EXPECT() *MockRepositoriesServiceMockRecorder {
	return m.recorder
}

// Edit mocks base method.
func (m *MockRepositoriesService) Edit(ctx context.Context, org, repo string, repository *github.Repository) (*github.Repository, *github.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Edit", ctx, org, repo, repository)
	ret0, _ := ret[0].(*github.Repository)
	ret1, _ := ret[1].(*github.Response)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Edit indicates an expected call of Edit.
func (mr *MockRepositoriesServiceMockRecorder) Edit(ctx, org, repo, repository interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Edit", reflect.TypeOf((*MockRepositoriesService)(nil).Edit), ctx, org, repo, repository)
}
