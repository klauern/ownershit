package ownershit

import (
	"errors"
	"testing"

	"go.uber.org/mock/gomock"
)

var ErrGenerated = errors.New("generated error")

func TestGitHubClient_SetDefaultBranch(t *testing.T) {
	mock := setupMocks(t)
	mock.repoMock.EXPECT().Edit(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil, nil)
	if err := mock.client.SetDefaultBranch(mock.client.Context, "klauern", "repo", "main"); err != nil {
		t.Errorf("did not expect an error here")
	}
	mock.repoMock.EXPECT().Edit(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil, ErrGenerated)
	if err := mock.client.SetDefaultBranch(mock.client.Context, "klauern", "repo", "main"); err == nil {
		t.Errorf("expected an error here")
	}
}

func TestGitHubClient_SetRepositoryDefaults(t *testing.T) {
	mock := setupMocks(t)
	mock.repoMock.EXPECT().Edit(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil, nil)
	if err := mock.client.SetRepositoryDefaults(mock.client.Context, "klauern", "ownershit", false, false, false); err != nil {
		t.Errorf("did not expect an error here")
	}
	mock.repoMock.EXPECT().Edit(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil, ErrGenerated)
	if err := mock.client.SetRepositoryDefaults(mock.client.Context, "klauern", "ownershit", false, false, false); err == nil {
		t.Errorf("expected an error here")
	}
}
