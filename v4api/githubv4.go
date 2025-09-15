package v4api

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

// LabelOperationError represents an error from a single label operation.
type LabelOperationError struct {
	Operation string // "create", "update", or "delete"
	LabelName string
	Error     error
}

// MultiLabelError represents multiple errors from label operations.
type MultiLabelError struct {
	Errors []LabelOperationError
}

func (e *MultiLabelError) Error() string {
	if len(e.Errors) == 0 {
		return "no errors"
	}
	if len(e.Errors) == 1 {
		err := e.Errors[0]
		return fmt.Sprintf("label operation failed: %s", err.Error.Error())
	}
	// For multiple errors, include count and first error's message for quicker debugging
	firstErr := e.Errors[0]
	return fmt.Sprintf("multiple label operations failed (%d errors): first error: %s", len(e.Errors), firstErr.Error.Error())
}

// GetDetailedErrors returns detailed error messages for each failed operation.
func (e *MultiLabelError) GetDetailedErrors() []string {
	var details []string
	for _, err := range e.Errors {
		details = append(details, fmt.Sprintf("%s label '%s': %v", err.Operation, err.LabelName, err.Error))
	}
	return details
}

// V4ClientDefaultPageSize defines the default page size for GraphQL queries.
const V4ClientDefaultPageSize = 100

// GetTeams returns all teams for the specified organization, handling pagination
// under the hood and returning a flattened list of team edges.
func (c *GitHubV4Client) GetTeams(organization string) (OrganizationTeams, error) {
	orgTeams := []GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdge{}
	initResp, err := GetTeams(c.Context, c.client, TeamOrder{
		Direction: OrderDirectionDesc,
		Field:     TeamOrderFieldName,
	}, V4ClientDefaultPageSize, "", organization)
	if err != nil {
		return nil, fmt.Errorf("failed to get teams (initial request): %w", err)
	}
	log.Debug().Interface("teams", initResp).Msg("initial teams")
	orgTeams = append(orgTeams, initResp.Organization.Teams.Edges...)
	hasNextPage := initResp.Organization.Teams.PageInfo.HasNextPage
	cursor := initResp.Organization.Teams.PageInfo.EndCursor
	for hasNextPage {
		log.Debug().Str("cursor", cursor).Msg("has next page")
		resp, err := GetTeams(c.Context, c.client, TeamOrder{
			Direction: OrderDirectionDesc,
			Field:     TeamOrderFieldName,
		}, V4ClientDefaultPageSize, cursor, organization)
		if err != nil {
			return nil, fmt.Errorf("failed to get teams (pagination request): %w", err)
		}
		hasNextPage = resp.Organization.Teams.PageInfo.HasNextPage
		cursor = resp.Organization.Teams.PageInfo.EndCursor
		orgTeams = append(orgTeams, resp.Organization.Teams.Edges...)
	}

	log.Debug().Interface("teams", orgTeams).Msg("teams")

	return orgTeams, nil
}

// GetRateLimit retrieves the API rate limit status for the current token and
// returns the response, logging details at debug level.
func (c *GitHubV4Client) GetRateLimit() (RateLimit, error) {
	resp, err := GetRateLimit(c.Context, c.client)
	if err != nil {
		return RateLimit{}, fmt.Errorf("can't get ratelimit: %w", err)
	}
	log.Debug().Interface("rate-limit", resp).Msg("rate limit")
	return RateLimit(*resp), nil
}

// fetchExistingLabels retrieves all existing labels from the repository.
func (c *GitHubV4Client) fetchExistingLabels(repo, owner string) (existingLabels map[string]Label, repoID string, err error) {
	existingLabels = map[string]Label{}
	cursor := ""

	for {
		labelResp, err := GetRepositoryIssueLabels(c.Context, c.client, repo, owner, cursor)
		if err != nil {
			return nil, "", fmt.Errorf("can't get labels for %s/%s: %w", owner, repo, err)
		}
		if repoID == "" {
			repoID = labelResp.Repository.Id
		}
		for i := range labelResp.Repository.Labels.Edges {
			label := labelResp.Repository.Labels.Edges[i]
			existingLabels[label.Node.Name] = Label(label.Node)
		}
		if !labelResp.Repository.Labels.PageInfo.HasNextPage {
			break
		}
		cursor = labelResp.Repository.Labels.PageInfo.EndCursor
	}

	return existingLabels, repoID, nil
}

// computeLabelDiff determines what labels need to be created, updated, or deleted.
func computeLabelDiff(desiredLabels []Label, existingLabels map[string]Label) (toCreate, toUpdate, toDelete []Label) {
	// Process desired labels to determine creates and updates
	for i := range desiredLabels {
		desiredLabel := &desiredLabels[i]
		if existingLabel, exists := existingLabels[desiredLabel.Name]; exists {
			// Check if update is needed by comparing fields
			if existingLabel.Color != desiredLabel.Color ||
				existingLabel.Description != desiredLabel.Description {
				// Propagate the ID from the existing label to the desired label
				desiredLabel.Id = existingLabel.Id
				toUpdate = append(toUpdate, *desiredLabel)
			}
			// Remove from map so remaining items are deletions
			delete(existingLabels, desiredLabel.Name)
		} else {
			// Label doesn't exist, needs to be created
			toCreate = append(toCreate, *desiredLabel)
		}
	}

	// Remaining items in existingLabels are deletions
	for name := range existingLabels {
		toDelete = append(toDelete, existingLabels[name])
	}

	return toCreate, toUpdate, toDelete
}

// executeLabelOperations performs create, update, and delete operations on labels.
func (c *GitHubV4Client) executeLabelOperations(repoID string, toCreate, toUpdate, toDelete []Label) []LabelOperationError {
	var errors []LabelOperationError

	// Execute creates
	for i := range toCreate {
		label := &toCreate[i]
		log.Debug().Interface("label", label).Msg("creating label")
		_, err := CreateLabel(c.Context, c.client, CreateLabelInput{
			Name:         label.Name,
			Color:        label.Color,
			Description:  label.Description,
			RepositoryId: repoID,
		})
		if err != nil {
			errors = append(errors, LabelOperationError{
				Operation: "create",
				LabelName: label.Name,
				Error:     err,
			})
		}
	}

	// Execute updates
	for i := range toUpdate {
		label := &toUpdate[i]
		log.Debug().Interface("label", label).Msg("updating label")
		_, err := UpdateLabel(c.Context, c.client, UpdateLabelInput{
			Name:        label.Name,
			Color:       label.Color,
			Description: label.Description,
			Id:          label.Id,
		})
		if err != nil {
			errors = append(errors, LabelOperationError{
				Operation: "update",
				LabelName: label.Name,
				Error:     err,
			})
		}
	}

	// Execute deletions
	for i := range toDelete {
		label := &toDelete[i]
		log.Debug().Interface("label", label).Msg("deleting label")
		_, err := DeleteLabel(c.Context, c.client, DeleteLabelInput{
			Id: label.Id,
		})
		if err != nil {
			errors = append(errors, LabelOperationError{
				Operation: "delete",
				LabelName: label.Name,
				Error:     err,
			})
		}
	}

	return errors
}

// SyncLabels ensures the repository's labels match the provided desired set by
// creating, updating, and deleting labels as needed using GraphQL mutations.
func (c *GitHubV4Client) SyncLabels(repo, owner string, labels []Label) error {
	// Fetch existing labels
	existingLabels, repoID, err := c.fetchExistingLabels(repo, owner)
	if err != nil {
		return err
	}

	// Compute diff: determine what needs to be created, updated, or deleted
	toCreate, toUpdate, toDelete := computeLabelDiff(labels, existingLabels)

	// Execute operations
	errors := c.executeLabelOperations(repoID, toCreate, toUpdate, toDelete)

	// Return aggregated error if any operations failed
	if len(errors) > 0 {
		return &MultiLabelError{Errors: errors}
	}
	return nil
}
