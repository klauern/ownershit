package ownershit

import (
	"github.com/rs/zerolog/log"
)

type PermissionsLevel string

const (
	Admin PermissionsLevel = "admin"
	Read  PermissionsLevel = "pull"
	Write PermissionsLevel = "push"
)

type Permissions struct {
	Team  string `yaml:"name"`
	ID    int64
	Level PermissionsLevel `yaml:"level"`
}

type PermissionsSettings struct {
	TeamPermissions []*Permissions `yaml:"team"`
	Repositories    []struct {
		Name      string
		Wiki      bool
		Issues    bool
		Projects  bool
		RepoPerms []*Permissions `yaml:"perms"`
	} `yaml:"repositories"`
	Organization string `yaml:"organization"`
}

func MapPermissions(settings *PermissionsSettings, client *GitHubClient) {
	for _, repo := range settings.Repositories {
		if len(settings.TeamPermissions) > 0 {
			for _, perm := range settings.TeamPermissions {
				log.Info().
					Interface("repository", repo.Name).
					Msg("Adding Permissions to repository")
				log.Debug().
					Interface("repository", repo.Name).
					Interface("permissions", perm).
					Msg("permissions to add to repository")
				err := client.AddPermissions(settings.Organization, repo.Name, *perm)
				if err != nil {
					log.Err(err).
						Interface("repository", repo.Name).
						Interface("permissions", perm).
						Msg("setting team permissions")
				}
			}
		}
		repoID, err := client.GetRepository(repo.Name, settings.Organization)
		if err != nil {
			log.Err(err).Str("repository", repo.Name).Msg("getting repository")
		} else {
			log.Debug().
				Interface("repoID", repoID).
				Msg("Repository ID")
			err := client.SetRepository(&repoID, repo.Wiki, repo.Issues, repo.Projects)
			if err != nil {
				log.Err(err).
					Interface("repoID", repoID).
					Bool("wikiEnabled", repo.Wiki).
					Bool("issuesEnabled", repo.Issues).
					Bool("projectsEnabled", repo.Projects).
					Msg("setting repository fields")
			}
		}
	}
}
