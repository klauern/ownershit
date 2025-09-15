package v4api

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/rs/zerolog/log"
)

var (
	// ErrMissingRepositoryID indicates the repository ID was absent in a GraphQL response.
	ErrMissingRepositoryID = errors.New("missing repository ID in response")
)

// LabelOperationError represents an error from a single label operation.
type LabelOperationError struct {
	Operation string // "create", "update", or "delete"
	LabelName string
	Err       error
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
		return fmt.Sprintf("label operation failed: %v", err.Err)
	}
	// For multiple errors, include count and first error's message for quicker debugging
	firstErr := e.Errors[0]
	return fmt.Sprintf("multiple label operations failed (%d errors): first error: %v", len(e.Errors), firstErr.Err)
}

// GetDetailedErrors returns detailed error messages for each failed operation.
func (e *MultiLabelError) GetDetailedErrors() []string {
	var details []string
	for _, err := range e.Errors {
		details = append(details, fmt.Sprintf("%s label '%s': %v", err.Operation, err.LabelName, err.Err))
	}
	return details
}

// Unwrap returns the underlying error(s) for compatibility with errors.Is/As.
func (e *MultiLabelError) Unwrap() error {
	if len(e.Errors) == 0 {
		return nil
	}
	if len(e.Errors) == 1 {
		return e.Errors[0].Err
	}
	errs := make([]error, 0, len(e.Errors))
	for _, le := range e.Errors {
		if le.Err != nil {
			errs = append(errs, le.Err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
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
	log.Debug().
		Int("edges", len(initResp.Organization.Teams.Edges)).
		Interface("pageInfo", initResp.Organization.Teams.PageInfo).
		Msg("initial teams page")
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

	log.Debug().Int("count", len(orgTeams)).Msg("teams fetched")

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
			if labelResp.Repository.Id == "" {
				return nil, "", fmt.Errorf("%w for %s/%s", ErrMissingRepositoryID, owner, repo)
			}
			repoID = labelResp.Repository.Id
		}
		for i := range labelResp.Repository.Labels.Edges {
			label := labelResp.Repository.Labels.Edges[i]
			key := strings.ToLower(label.Node.Name)
			existingLabels[key] = Label(label.Node)
		}
		if !labelResp.Repository.Labels.PageInfo.HasNextPage {
			break
		}
		cursor = labelResp.Repository.Labels.PageInfo.EndCursor
	}

	return existingLabels, repoID, nil
}

// computeLabelDiff determines what labels need to be created, updated, or deleted.
func computeLabelDiff(desiredLabels []Label, existingLabels map[string]Label) (toCreate, toUpdate []Label, toDeleteIDs []string) {
	// Process desired labels to determine creates and updates
	for i := range desiredLabels {
		desiredLabel := &desiredLabels[i]
		key := strings.ToLower(desiredLabel.Name)
		if existingLabel, exists := existingLabels[key]; exists {
			// Check if update is needed by comparing fields
			if existingLabel.Color != desiredLabel.Color ||
				existingLabel.Description != desiredLabel.Description ||
				existingLabel.Name != desiredLabel.Name {
				// Build update without mutating the input.
				toUpdate = append(toUpdate, Label{
					Id:          existingLabel.Id,
					Name:        desiredLabel.Name,
					Color:       desiredLabel.Color,
					Description: desiredLabel.Description,
				})
			}
			// Remove from map so remaining items are deletions
			delete(existingLabels, key)
		} else {
			// Label doesn't exist, needs to be created; avoid copying extra fields
			toCreate = append(toCreate, Label{
				Name:        desiredLabel.Name,
				Color:       desiredLabel.Color,
				Description: desiredLabel.Description,
			})
		}
	}

	// Remaining items in existingLabels are deletions; collect deterministically
	names := make([]string, 0, len(existingLabels))
	for name := range existingLabels {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		toDeleteIDs = append(toDeleteIDs, existingLabels[name].Id)
	}

	return toCreate, toUpdate, toDeleteIDs
}

// executeLabelOperations performs create, update, and delete operations on labels.
func (c *GitHubV4Client) executeLabelOperations(repoID string, toCreate, toUpdate []Label, toDeleteIDs []string) []LabelOperationError {
	var opErrors []LabelOperationError

	// Execute creates
	for i := range toCreate {
		label := &toCreate[i]
		log.Debug().Interface("label", label).Msg("creating label")
		_, err := CreateLabel(c.Context, c.client, CreateLabelInput{
			ClientMutationId: "create-label-" + label.Name,
			Name:             label.Name,
			Color:            label.Color,
			Description:      label.Description,
			RepositoryId:     repoID,
		})
		if err != nil {
			opErrors = append(opErrors, LabelOperationError{
				Operation: "create",
				LabelName: label.Name,
				Err:       err,
			})
		}
	}

	// Execute updates
	for i := range toUpdate {
		label := &toUpdate[i]
		log.Debug().Interface("label", label).Msg("updating label")
		_, err := UpdateLabel(c.Context, c.client, UpdateLabelInput{
			ClientMutationId: "update-label-" + label.Id,
			Name:             label.Name,
			Color:            label.Color,
			Description:      label.Description,
			Id:               label.Id,
		})
		if err != nil {
			opErrors = append(opErrors, LabelOperationError{
				Operation: "update",
				LabelName: label.Name,
				Err:       err,
			})
		}
	}

	// Execute deletions
	for _, id := range toDeleteIDs {
		log.Debug().Str("label_id", id).Msg("deleting label")
		_, err := DeleteLabel(c.Context, c.client, DeleteLabelInput{
			ClientMutationId: "delete-label-" + id,
			Id:               id,
		})
		if err != nil {
			opErrors = append(opErrors, LabelOperationError{
				Operation: "delete",
				LabelName: fmt.Sprintf("id=%s", id),
				Err:       err,
			})
		}
	}

	return opErrors
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
	toCreate, toUpdate, toDeleteIDs := computeLabelDiff(labels, existingLabels)

	// Plan summary for visibility
	log.Debug().
		Int("create", len(toCreate)).
		Int("update", len(toUpdate)).
		Int("delete", len(toDeleteIDs)).
		Msg("label sync plan")

		// Execute operations
	opErrors := c.executeLabelOperations(repoID, toCreate, toUpdate, toDeleteIDs)

	// Return aggregated error if any operations failed
	if len(opErrors) > 0 {
		return &MultiLabelError{Errors: opErrors}
	}
	return nil
}
