package ownershit

import (
	"testing"
)

func TestMapPermissions(t *testing.T) {
	type args struct {
		settings *PermissionsSettings
		client   *GitHubClient
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			MapPermissions(tt.args.settings, tt.args.client)
		})
	}
}

func TestUpdateBranchMergeStrategies(t *testing.T) {
	type args struct {
		settings *PermissionsSettings
		client   *GitHubClient
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			UpdateBranchMergeStrategies(tt.args.settings, tt.args.client)
		})
	}
}
