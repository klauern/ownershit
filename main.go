package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"

	"github.com/google/go-github/v29/github"
	cli "github.com/urfave/cli/v2"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
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
	Level PermissionsLevel `yaml:"level`
}

type PermissionsSettings struct {
	TeamPermissions []*Permissions `yaml:"team"`
	Repositories    []string       `yaml:"repositories"`
	Organization    string         `yaml:"organization"`
}

var repositoriesYAMLConfig string

func main() {

	app := &cli.App{
		Name:  "ownershit",
		Usage: "fix up team ownership of your repositories in an organization",
		UsageText: "ownershit --config repositories.yaml",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "config",
				Value:       "repositories.yaml",
				Usage:       "configuration of repository updates to perform",
				Destination: &repositoriesYAMLConfig,
			},
		},
		Action: runApp,
		Authors: []*cli.Author{{Name:  "Nick Klauer", Email: "klauer@gmail.com",}},
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

	var wg sync.WaitGroup
	for _, team := range settings.TeamPermissions {
		wg.Add(1)
		go getTeamSlug(&wg, client, ctx, settings, team)
	}
	wg.Wait()

	for _, repo := range settings.Repositories {
		if len(settings.TeamPermissions) > 0 {
			for _, perm := range settings.TeamPermissions {
				fmt.Printf("Repo: %v, Permission: %+v\n", repo, perm)
				go addPermissions(client, ctx, repo, settings.Organization, *perm)
			}
		}
	}

	return nil
}

func getTeamSlug(wg *sync.WaitGroup, client *github.Client, ctx context.Context, settings *PermissionsSettings, team *Permissions) error {
	defer wg.Done()
	t, _, err := client.Teams.GetTeamBySlug(ctx, settings.Organization, team.Team)
	if err != nil {
		return fmt.Errorf("Unable to get Team from organization: %w", err)
	}
	fmt.Printf("Team: %v ID: %v\n", team.Team, *t.ID)
	team.ID = *t.ID
	return nil
}

