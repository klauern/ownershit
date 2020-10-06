package ownershit

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-github/v32/github"
	"github.com/shurcooL/githubv4"
)

func TestGitHubClient_QueryArchivableIssues(t *testing.T) {
	// mocks := setupMocks(t)
	// mocks.graphMock.EXPECT().Query(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Do()
	type fields struct {
		Teams        TeamsService
		Repositories RepositoriesService
		Graph        GraphQLClient
		V3           *github.Client
		V4           *githubv4.Client
		Context      context.Context
	}
	type args struct {
		username string
		forks    int
		stars    int
		maxDays  int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []RepositoryInfo
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := defaultGitHubClient()
			got, err := c.QueryArchivableIssues(tt.args.username, tt.args.forks, tt.args.stars, tt.args.maxDays)
			if (err != nil) != tt.wantErr {
				t.Errorf("GitHubClient.QueryArchivableIssues() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GitHubClient.QueryArchivableIssues() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGitHubClient_MutateArchiveRepository(t *testing.T) {
	// mocks := setupMocks()
	type fields struct {
		Teams        TeamsService
		Repositories RepositoriesService
		Graph        GraphQLClient
		V3           *github.Client
		V4           *githubv4.Client
		Context      context.Context
	}
	type args struct {
		repo RepositoryInfo
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := defaultGitHubClient()
			if err := c.MutateArchiveRepository(tt.args.repo); (err != nil) != tt.wantErr {
				t.Errorf("GitHubClient.MutateArchiveRepository() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_removeElement(t *testing.T) {
	type args struct {
		slice []RepositoryInfo
		s     int
	}
	tests := []struct {
		name string
		args args
		want []RepositoryInfo
	}{
		{
			name: "remove second element",
			args: args{
				slice: []RepositoryInfo{
					{
						ID: githubv4.String("0"),
					},
					{
						ID: githubv4.String("1"),
					},
					{
						ID: githubv4.String("2"),
					},
				},
				s: 1,
			},
			want: []RepositoryInfo{
				{
					ID: githubv4.String("0"),
				},
				{
					ID: githubv4.String("2"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := removeElement(tt.args.slice, tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("removeElement() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepositoryInfos_Len(t *testing.T) {
	tests := []struct {
		name string
		r    RepositoryInfos
		want int
	}{
		{
			name: "three",
			r: RepositoryInfos{
				RepositoryInfo{},
				RepositoryInfo{},
				RepositoryInfo{},
			},
			want: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.Len(); got != tt.want {
				t.Errorf("RepositoryInfos.Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepositoryInfos_Swap(t *testing.T) {
	type args struct {
		i int
		j int
	}
	tests := []struct {
		name string
		r    RepositoryInfos
		args args
	}{
		{
			name: "test swap",
			args: args{
				i: 1,
				j: 0,
			},
			r: RepositoryInfos{
				RepositoryInfo{
					Name: githubv4.String("first"),
				},
				RepositoryInfo{
					Name: githubv4.String("second"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.r.Swap(tt.args.i, tt.args.j)
			fmt.Println(tt.r[1].Name)
			if string(tt.r[1].Name) != "first" {
				t.Fail()
			}
		})
	}
}

func TestReposByName_Less(t *testing.T) {
	type fields struct {
		RepositoryInfos RepositoryInfos
	}
	type args struct {
		i int
		j int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "",
			args: args{
				i: 0,
				j: 1,
			},
			fields: fields{
				RepositoryInfos: RepositoryInfos{
					RepositoryInfo{
						Name: githubv4.String("xylophone"),
					},
					RepositoryInfo{
						Name: githubv4.String("alphabet"),
					},
				},
			},
			want: false,
		},
		{
			name: "",
			args: args{
				i: 1,
				j: 0,
			},
			fields: fields{
				RepositoryInfos: RepositoryInfos{
					RepositoryInfo{
						Name: githubv4.String("xylophone"),
					},
					RepositoryInfo{
						Name: githubv4.String("alphabet"),
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := ReposByName{
				RepositoryInfos: tt.fields.RepositoryInfos,
			}
			if got := r.Less(tt.args.i, tt.args.j); got != tt.want {
				t.Errorf("ReposByName.Less() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSortedRepositoryInfo(t *testing.T) {
	type args struct {
		repos []RepositoryInfo
	}
	tests := []struct {
		name string
		args args
		want []RepositoryInfo
	}{
		{
			name: "sorted",
			args: args{
				repos: []RepositoryInfo{
					{
						Name: githubv4.String("x"),
					},
					{
						Name: githubv4.String("z"),
					},
					{
						Name: githubv4.String("y"),
					},
				},
			},
			want: []RepositoryInfo{
				{
					Name: githubv4.String("x"),
				},
				{
					Name: githubv4.String("y"),
				},
				{
					Name: githubv4.String("z"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SortedRepositoryInfo(tt.args.repos); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SortedRepositoryInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}
