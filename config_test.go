package ownershit

import "testing"

func TestMapPermissions(t *testing.T) {
	type args struct {
		settings *PermissionsSettings
		err      error
		client   *GitHubClient
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "default",
			args: args{
				settings: &PermissionsSettings{},
				client:   nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			MapPermissions(tt.args.settings, tt.args.client)
		})
	}
}
