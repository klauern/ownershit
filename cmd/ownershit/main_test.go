package main

import (
	"context"
	"testing"

	"github.com/urfave/cli/v2"
)

func Test_main(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			main()
		})
	}
}

func Test_readConfigs(t *testing.T) {
	type args struct {
		c *cli.Context
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := readConfigs(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("readConfigs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_syncCommand(t *testing.T) {
	type args struct {
		c *cli.Context
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := syncCommand(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("syncCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_branchCommand(t *testing.T) {
	type args struct {
		c *cli.Context
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := branchCommand(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("branchCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSetGithubClient(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetGithubClient(tt.args.ctx)
		})
	}
}
