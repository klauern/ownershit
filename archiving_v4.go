package ownershit

import (
	"sort"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/shurcooL/githubv4"
)

/*
query archivableRepositories {
  search(query: $user, type: REPOSITORY, first: 10, after: $afeerthing) {
    pageInfo {
      hasNextPage
      startCursor
      endCursor
    }
    repositoryCount
    nodes {
      ... on Repository {
        id
        name
        isFork
        forkCount
		isArchived
		stargazersCount
		updatedAt
		watchers	{
			totalCount
		}
      }
    }
  }
}
*/
type ArchivableIssuesQuery struct {
	Search struct {
		PageInfo struct {
			HasNextPage githubv4.Boolean
			EndCursor   githubv4.String
			StartCursor githubv4.String
		}
		RepositoryCount int
		Repos           []struct {
			Repository RepositoryInfo `graphql:"... on Repository"`
		} `graphql:"nodes"`
	} `graphql:"search(query: $user, type: REPOSITORY, first: $first, after: $repositoryCursor)"`
}

type RepositoryInfo struct {
	ID             githubv4.String
	Name           githubv4.String
	ForkCount      githubv4.Int
	IsArchived     githubv4.Boolean
	IsFork         githubv4.Boolean
	StargazerCount githubv4.Int
	UpdatedAt      githubv4.DateTime
	Watchers       struct {
		TotalCount githubv4.Int
	}
}

const PerPage = 100
const OneDay = time.Hour * 24

func (r *RepositoryInfo) IsArchivable(forks, stars, maxDays, maxWatchers int) bool {
	log.Debug().Fields(map[string]interface{}{
		"isArchived":  r.IsArchived,
		"isFork":      r.IsFork,
		"forks":       r.ForkCount,
		"stars":       r.StargazerCount,
		"lastUpdated": r.UpdatedAt,
		"watchers":    r.Watchers.TotalCount,
	})
	if r.IsArchived {
		return true
	}
	if r.IsFork {
		return true
	}
	if int(r.ForkCount) > forks {
		return true
	}
	if int(r.StargazerCount) > stars {
		return true
	}
	if r.UpdatedAt.Time.After(time.Now().Add(-time.Duration(maxDays) * OneDay)) {
		return true
	}
	if int(r.Watchers.TotalCount) > maxWatchers {
		return true
	}
	return false
}

func (c *GitHubClient) QueryArchivableIssues(username string, forks, stars, maxDays, maxWatchers int) ([]RepositoryInfo, error) {
	var query ArchivableIssuesQuery
	variables := map[string]interface{}{
		"user":             githubv4.String("user:" + username),
		"first":            githubv4.Int(PerPage),
		"repositoryCursor": (*githubv4.String)(nil),
	}
	var repos []RepositoryInfo
	for {
		err := c.Graph.Query(c.Context, &query, variables)
		if err != nil {
			log.Err(err).
				Str("username", username).
				Msg("error querying for archivable repositories")
			return nil, err
		}

		for _, v := range query.Search.Repos {
			repos = append(repos, v.Repository)
		}
		if !query.Search.PageInfo.HasNextPage {
			break
		}
		variables["repositoryCursor"] = githubv4.NewString(query.Search.PageInfo.EndCursor)
	}
	for i := 0; i < len(repos); i++ {
		if repos[i].IsArchivable(forks, stars, maxDays, maxWatchers) {
			repos = removeElement(repos, i)
			i--
			continue
		}
	}
	return repos, nil
}

func (c *GitHubClient) MutateArchiveRepository(repo RepositoryInfo) error {
	var mutation struct {
		ArchiveRepository struct {
			Repository struct {
				ID            githubv4.String
				NameWithOwner githubv4.String
			}
		} `graphql:"archiveRepository(input: $input)"`
	}
	input := githubv4.ArchiveRepositoryInput{
		RepositoryID: repo.ID,
	}
	err := c.V4.Mutate(c.Context, &mutation, input, nil)
	if err != nil {
		log.Err(err).Interface("repository", repo.Name).Msg("Unable to archive repository")
		return err
	}
	return nil
}

func removeElement(slice []RepositoryInfo, s int) []RepositoryInfo {
	return append(slice[:s], slice[s+1:]...)
}

// Everything behind this is related to sorting a list of RepositoryInfo.
type RepositoryInfos []RepositoryInfo
type ReposByName struct{ RepositoryInfos }

func (r RepositoryInfos) Len() int      { return len(r) }
func (r RepositoryInfos) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r ReposByName) Less(i, j int) bool {
	return string(r.RepositoryInfos[i].Name) < string(r.RepositoryInfos[j].Name)
}
func SortedRepositoryInfo(repos []RepositoryInfo) []RepositoryInfo {
	sort.Sort(ReposByName{repos})
	return repos
}
