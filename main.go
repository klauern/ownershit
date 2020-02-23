package main

import (
	"context"
	"fmt"
	"github.com/google/go-github/v29/github"
	"github.com/shurcooL/githubv4"
	"github.com/urfave/cli/v2"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"os"
)

var GithubToken = os.Getenv("GITHUB_TOKEN")

type PermissionsLevel string

const (
	PermissionsAdmin PermissionsLevel = "admin"
	PermissionsRead  PermissionsLevel = "pull"
	PermissionsWrite PermissionsLevel = "push"
)

type Permissions struct {
	Team  string `yaml:"name"`
	ID    int64
	Level PermissionsLevel `yaml:"level"`
}

type PermissionsSettings struct {
	TeamPermissions []*Permissions `yaml:"team"`
	Repositories    []struct {
		Name     string
		Wiki     bool
		Issues   bool
		Projects bool
	} `yaml:"repositories"`
	Organization string `yaml:"organization"`
}

var repositoriesYAMLConfig string

func main() {

	app := &cli.App{
		Name:      "ownershit",
		Usage:     "fix up team ownership of your repositories in an organization",
		UsageText: "ownershit --config repositories.yaml",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "config",
				Value:       "repositories.yaml",
				Usage:       "configuration of repository updates to perform",
				Destination: &repositoriesYAMLConfig,
			},
		},
		Action:  runApp,
		Authors: []*cli.Author{{Name: "Nick Klauer", Email: "klauer@gmail.com",}},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}

func addPermissions(client *github.Client, ctx context.Context, repo string, organization string, perm Permissions) error {
	resp, err := client.Teams.AddTeamRepoBySlug(ctx, organization, perm.Team, organization, repo, &github.TeamAddTeamRepoOptions{Permission: string(perm.Level)})
	if err != nil {
		fmt.Printf("error adding %v as collaborator to %v: %v\n", perm.Team, repo, resp.Status)
		resp, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("Unable to read response body: %w", err)
		}
		fmt.Println(string(resp))
	}
	return nil
}

func runApp(c *cli.Context) error {
	file, err := ioutil.ReadFile(repositoriesYAMLConfig)
	if err != nil {
		return fmt.Errorf("config file error: %w", err)
	}

	settings := &PermissionsSettings{}
	if err = yaml.Unmarshal(file, settings); err != nil {
		return fmt.Errorf("config file yaml unmarshal error: %w", err)
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: GithubToken})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	ghv4Client := githubv4.NewClient(tc)

	for _, team := range settings.TeamPermissions {
		err = getTeamSlug(client, ctx, settings, team)
		if err != nil {
			fmt.Println(err)
		}
	}

	for _, repo := range settings.Repositories {
		if len(settings.TeamPermissions) > 0 {
			for _, perm := range settings.TeamPermissions {
				fmt.Printf("Repo: %v, Permission: %+v\n", repo.Name, perm)
				err = addPermissions(client, ctx, repo.Name, settings.Organization, *perm)
				if err != nil {
					fmt.Println(err)
				}
			}
		}
		repoID, err := GetRepository(ghv4Client, ctx, repo.Name, settings.Organization)
		if err != nil {
			fmt.Println(err)
		} else {
			err := SetRepository(ghv4Client, ctx, repoID, repo.Wiki, repo.Issues, repo.Projects)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}

func getTeamSlug(client *github.Client, ctx context.Context, settings *PermissionsSettings, team *Permissions) error {
	t, _, err := client.Teams.GetTeamBySlug(ctx, settings.Organization, team.Team)
	if err != nil {
		return fmt.Errorf("Unable to get Team from organization: %w", err)
	}
	fmt.Printf("Team: %v ID: %v\n", team.Team, *t.ID)
	team.ID = *t.ID
	return nil
}

func GetRepository(client *githubv4.Client, ctx context.Context, name, owner string) (*githubv4.ID, error) {
	var query struct {
		Repository struct {
			ID                 githubv4.ID
			HasWikiEnabled     githubv4.Boolean
			HasIssuesEnabled   githubv4.Boolean
			HasProjectsEnabled githubv4.Boolean
		} `graphql:"repository(owner: $owner, name: $name)"`
	}
	err := client.Query(ctx, &query, map[string]interface{}{
		"owner": githubv4.String(owner),
		"name":  githubv4.String(name),
	})
	if err != nil {
		fmt.Printf("error retrieving %s/%s: %v\n", owner, name, err)
		return nil, err
	}
	fmt.Printf("repository %s/%s: wiki:%v,issues:%v,projects:%v\n", owner, name, query.Repository.HasWikiEnabled, query.Repository.HasIssuesEnabled, query.Repository.HasProjectsEnabled)
	return &query.Repository.ID, nil
}

func SetRepository(client *githubv4.Client, ctx context.Context, id githubv4.ID, wiki, issues, project bool) error {
	var mutation struct {
		UpdateRepository struct {
			ClientMutationID githubv4.String
			Repository struct {
				ID githubv4.ID
				HasWikiEnabled githubv4.Boolean
				HasIssuesEnabled githubv4.Boolean
				HasProjectsEnabled githubv4.Boolean
			}
		} `graphql:"updateRepository(input: $input)"`
	}
	input := githubv4.UpdateRepositoryInput{
		RepositoryID:       id,
		HasWikiEnabled:     githubv4.NewBoolean(githubv4.Boolean(wiki)),
		HasIssuesEnabled:   githubv4.NewBoolean(githubv4.Boolean(issues)),
		HasProjectsEnabled: githubv4.NewBoolean(githubv4.Boolean(project)),
	}
	err := client.Mutate(ctx, &mutation, input, nil)
	if err != nil {
		fmt.Printf("error updating repository ID %v: %v", id, err)
		return err
	}
	return nil
}
