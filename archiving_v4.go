package ownershit

import (
	"sort"

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
	ID         githubv4.String
	Name       githubv4.String
	ForkCount  githubv4.Int
	IsArchived githubv4.Boolean
	IsFork     githubv4.Boolean
}

func (c *GitHubClient) QueryArchivableIssues(username string, forks, stars int) ([]RepositoryInfo, error) {
	var query ArchivableIssuesQuery
	variables := map[string]interface{}{
		"user":             githubv4.String("user:" + username),
		"first":            githubv4.Int(100),
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
	for i, v := range repos {
		if v.IsArchived {
			repos = removeElement(repos, i)
			continue
		}
		if v.IsFork {
			repos = removeElement(repos, i)
			continue
		}
		if int(v.ForkCount) >= forks {
			repos = removeElement(repos, i)
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

type RepositoryInfos []RepositoryInfo

func (r RepositoryInfos) Len() int      { return len(r) }
func (r RepositoryInfos) Swap(i, j int) { r[i], r[j] = r[j], r[i] }

type ReposByName struct{ RepositoryInfos }

func (r ReposByName) Less(i, j int) bool {
	return string(r.RepositoryInfos[i].Name) < string(r.RepositoryInfos[j].Name)
}

func SortedRepositoryInfo(repos []RepositoryInfo) []RepositoryInfo {
	sort.Sort(ReposByName{repos})
	return repos
}
