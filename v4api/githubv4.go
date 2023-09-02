package v4api

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

const V4ClientDefaultPageSize = 100

func (c *GitHubV4Client) GetTeams() (OrganizationTeams, error) {
	orgTeams := []GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdge{}
	initResp, err := GetTeams(c.Context, c.client, TeamOrder{
		Direction: OrderDirectionDesc,
		Field:     TeamOrderFieldName,
	}, V4ClientDefaultPageSize, "")
	if err != nil {
		panic(err)
	}
	log.Info().Interface("teams", initResp).Msg("initial teams")
	hasNextPage := initResp.Organization.Teams.PageInfo.HasNextPage
	cursor := initResp.Organization.Teams.PageInfo.EndCursor
	for hasNextPage {
		log.Debug().Str("cursor", cursor).Msg("has next page")
		resp, err := GetTeams(c.Context, c.client, TeamOrder{
			Direction: OrderDirectionDesc,
			Field:     TeamOrderFieldName,
		}, V4ClientDefaultPageSize, cursor)
		if err != nil {
			panic(err)
		}
		hasNextPage = resp.Organization.Teams.PageInfo.HasNextPage
		cursor = resp.Organization.Teams.PageInfo.EndCursor
		orgTeams = append(orgTeams, resp.Organization.Teams.Edges...)
	}

	log.Debug().Interface("teams", orgTeams).Msg("teams")

	return orgTeams, nil
}

func (c *GitHubV4Client) GetRateLimit() (RateLimit, error) {
	resp, err := GetRateLimit(c.Context, c.client)
	if err != nil {
		return RateLimit{}, fmt.Errorf("can't get ratelimit: %w", err)
	}
	log.Debug().Interface("rate-limit", resp)
	return RateLimit(*resp), nil
}

func (c *GitHubV4Client) SyncLabels(repo string, labels []Label) error {
	labelsMap := map[string]Label{}
	labelResp, err := GetRepositoryIssueLabels(c.Context, c.client, repo, "zendesk", "")
	repoID := labelResp.Repository.Id
	if err != nil {
		return fmt.Errorf("can't get labels: %w", err)
	}
	for _, label := range labelResp.Repository.Labels.Edges {
		labelsMap[label.Node.Name] = Label(label.Node)
	}
	for _, label := range labels {
		// create some kind of type to track which are new, missing, or updated
		l, ok := labelsMap[label.Name]
		if ok {
			// Label name exists, so we need to update it
			log.Debug().Interface("label", label).Msg("label exists")
			_, err := UpdateLabel(c.Context, c.client, UpdateLabelInput{
				Name:        l.Name,
				Color:       l.Color,
				Description: l.Description,
				Id:          l.Id,
			})
			if err != nil {
				return fmt.Errorf("can't update label: %w", err)
			}
			delete(labelsMap, label.Name)
		} else {
			// Label name doesn't exist, so we need to create it
			log.Debug().Interface("label", label).Msg("label does not exist")
			_, err := CreateLabel(c.Context, c.client, CreateLabelInput{
				Name:         l.Name,
				Color:        l.Color,
				Description:  l.Description,
				RepositoryId: repoID,
			})
			if err != nil {
				return fmt.Errorf("can't create label: %w", err)
			}
			delete(labelsMap, label.Name)
		}
	}
	// now we have a map of labels that we need to delete
	for _, label := range labelsMap {
		_, err := DeleteLabel(c.Context, c.client, DeleteLabelInput{
			Id: label.Id,
		})
		if err != nil {
			return fmt.Errorf("can't delete label: %w", err)
		}
	}
	return nil
}
