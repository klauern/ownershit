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
