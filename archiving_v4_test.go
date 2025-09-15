package ownershit

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/shurcooL/githubv4"
	"go.uber.org/mock/gomock"
)

var (
	ErrTestError  = errors.New("test error here")
	ErrDummyError = errors.New("dummy error")
)

const oneDay = time.Hour * 24

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

type archivableTestFields struct {
	ID             githubv4.String
	Name           githubv4.String
	ForkCount      githubv4.Int
	IsArchived     githubv4.Boolean
	IsFork         githubv4.Boolean
	StargazerCount githubv4.Int
	UpdatedAt      githubv4.DateTime
	Watchers       struct{ TotalCount githubv4.Int }
}

type archivableTestArgs struct {
	forks    int
	stars    int
	maxDays  int
	watchers int
}

func getArchivableTestCases() []struct {
	name   string
	fields archivableTestFields
	args   archivableTestArgs
	want   bool
} {
	return []struct {
		name   string
		fields archivableTestFields
		args   archivableTestArgs
		want   bool
	}{
		{
			name:   "with stars",
			args:   archivableTestArgs{},
			fields: archivableTestFields{StargazerCount: githubv4.Int(1)},
			want:   false,
		},
		{
			name:   "with forks",
			args:   archivableTestArgs{},
			fields: archivableTestFields{ForkCount: githubv4.Int(1)},
			want:   false,
		},
		{
			name:   "with days",
			args:   archivableTestArgs{maxDays: 1},
			fields: archivableTestFields{UpdatedAt: githubv4.DateTime{Time: time.Now()}},
			want:   false,
		},
		{
			name:   "not with stars",
			args:   archivableTestArgs{stars: 1},
			fields: archivableTestFields{StargazerCount: githubv4.Int(1)},
			want:   false,
		},
		{
			name:   "not with forks",
			args:   archivableTestArgs{forks: 1},
			fields: archivableTestFields{ForkCount: githubv4.Int(1)},
			want:   false,
		},
		{
			name:   "not with days",
			args:   archivableTestArgs{maxDays: 1, forks: 10, stars: 10, watchers: 10},
			fields: archivableTestFields{UpdatedAt: githubv4.DateTime{Time: time.Now().Add(-oneDay)}},
			want:   true,
		},
		{
			name:   "already archived",
			args:   archivableTestArgs{},
			fields: archivableTestFields{IsArchived: githubv4.Boolean(true)},
			want:   false,
		},
		{
			name:   "is forked",
			args:   archivableTestArgs{},
			fields: archivableTestFields{IsFork: githubv4.Boolean(true)},
			want:   false,
		},
		{
			name:   "has watchers",
			args:   archivableTestArgs{},
			fields: archivableTestFields{Watchers: struct{ TotalCount githubv4.Int }{githubv4.Int(1)}},
			want:   false,
		},
		{
			name:   "not with watchers",
			args:   archivableTestArgs{watchers: 1},
			fields: archivableTestFields{Watchers: struct{ TotalCount githubv4.Int }{githubv4.Int(1)}},
			want:   false,
		},
	}
}

func TestRepositoryInfo_IsArchivable(t *testing.T) {
	tests := getArchivableTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := createRepositoryInfo(tt.fields)
			fmt.Println(tt.fields.UpdatedAt.Date())
			if got := r.IsArchivable(tt.args.forks, tt.args.stars, tt.args.maxDays, tt.args.watchers); got != tt.want {
				t.Errorf("RepositoryInfo.IsArchivable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func createRepositoryInfo(fields archivableTestFields) *RepositoryInfo {
	return &RepositoryInfo{
		ID:             fields.ID,
		Name:           fields.Name,
		ForkCount:      fields.ForkCount,
		IsArchived:     fields.IsArchived,
		IsFork:         fields.IsFork,
		StargazerCount: fields.StargazerCount,
		UpdatedAt:      fields.UpdatedAt,
		Watchers:       fields.Watchers,
	}
}

func TestQueryArchivableRepos(t *testing.T) {
	mocks := setupMocks(t)
	dummyVars := map[string]interface{}{
		"user":             githubv4.String("user:klauern"),
		"first":            githubv4.Int(100),
		"repositoryCursor": (*githubv4.String)(nil),
	}
	mocks.graphMock.EXPECT().Query(gomock.Any(), gomock.Any(), gomock.Eq(dummyVars)).Return(nil).Do(
		func(c context.Context, y *ArchivableRepositoriesQuery, v map[string]interface{}) {
			y.Search.Repos = []struct {
				Repository RepositoryInfo `graphql:"... on Repository"`
			}{{Repository: RepositoryInfo{IsArchived: githubv4.Boolean(true)}}}
			y.Search.RepositoryCount = 1
			y.Search.PageInfo = pageInfo{HasNextPage: true, EndCursor: githubv4.String("dummycursor")}
		})
	mocks.graphMock.EXPECT().Query(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Do(
		func(c context.Context, y *ArchivableRepositoriesQuery, v map[string]interface{}) {
			y.Search.Repos = []struct {
				Repository RepositoryInfo `graphql:"... on Repository"`
			}{}
			y.Search.RepositoryCount = 0
			y.Search.PageInfo = pageInfo{}
		})
	info, err := mocks.client.QueryArchivableRepos("klauern", 1, 1, 1, 1)
	if err != nil {
		t.Error("error unexpected")
	}
	if len(info) != 0 {
		t.Errorf("length of RepositoryInfo: %v, expected to be %v", len(info), 0)
	}
	mocks.graphMock.EXPECT().Query(gomock.Any(), gomock.Any(), gomock.Eq(dummyVars)).Return(ErrTestError)
	_, err = mocks.client.QueryArchivableRepos("klauern", 1, 1, 1, 1)
	if err == nil {
		t.Error("error expected")
	}
}

func TestMutateArchiveRepository(t *testing.T) {
	mock := setupMocks(t)
	dummyRepo := RepositoryInfo{
		Name: githubv4.String("hello"),
		ID:   githubv4.String("dummyID"),
	}
	mock.graphMock.EXPECT().Mutate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Nil()).Return(nil)
	if err := mock.client.MutateArchiveRepository(dummyRepo); err != nil {
		t.Error("did not expect error here")
	}
	mock.graphMock.EXPECT().Mutate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Nil()).Return(ErrDummyError)
	if err := mock.client.MutateArchiveRepository(dummyRepo); err == nil {
		t.Error("expected error here")
	}
}
