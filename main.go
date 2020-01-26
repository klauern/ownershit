package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/go-github/v29/github"
	"golang.org/x/oauth2"
	yaml "gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
)

var GithubToken = os.Getenv("GITHUB_TOKEN")

const DefaultOrg= "zendesk"

const (
	AffiliationAll = "all"
	AffiliationDirect = "direct"
	AffiliationOutside = "outside"
)

type PermissionsLevel string
const (
	PermissionsAdmin PermissionsLevel = "admin"
	PermissionsPull PermissionsLevel = "pull"
	PermissionsPush PermissionsLevel = "push"
)

type Permissions struct {
	Team  string `yaml:"name"`
	ID    int64
	Level PermissionsLevel `yaml:"level`
}

type PermissionsSettings struct {
	TeamPermissions []*Permissions `yaml:"team"`
	Repositories    []string      `yaml:"repositories"`
}

var repoConfigurationPath string

func init() {
	setFlags()
}

func setFlags() {
	flag.StringVar(&repoConfigurationPath, "config", "repositories.yaml", "configuration of repository updates to perform")
}


func main() {
	flag.Parse()

	file, err := ioutil.ReadFile(repoConfigurationPath)
	if err != nil {
		panic(err)
	}

	settings := &PermissionsSettings{}
	//fmt.Println(string(file))
	if err = yaml.Unmarshal(file, settings); err != nil {
		panic(err)
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: GithubToken})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	for _, team := range settings.TeamPermissions {
		t, _, err := client.Teams.GetTeamBySlug(ctx, DefaultOrg, team.Team)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Team ID: %v\n", *t.ID)
		team.ID = *t.ID

	}


	for _, repo := range settings.Repositories {
		if len(settings.TeamPermissions) > 0 {
			for _, perm := range settings.TeamPermissions {
				fmt.Printf("Repo: %v, Permission: %+v\n", repo, perm)
				addPermissions(client, ctx, repo, *perm)
			}
		}
	}
}

func addPermissions(client *github.Client, ctx context.Context, repo string, perm Permissions) {
	resp, err := client.Teams.AddTeamRepo(ctx, perm.ID, "zendesk", repo, &github.TeamAddTeamRepoOptions{Permission: string(perm.Level)})
	if err != nil {
		fmt.Printf("error adding %v as collaborator to %v: %v\n", perm.Team, repo, resp.Status)
		resp, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)

		}
		fmt.Println(string(resp))
	}
}
