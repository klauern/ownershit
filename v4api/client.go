package v4api

import (
	"context"
	"fmt"
	gLog "log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/rs/zerolog/log"
)

type GitHubV4Client struct {
	baseClient  *http.Client
	retryClient *retryablehttp.Client
	client      graphql.Client
	Context     context.Context
}

const (
	TIMEOUT_SECONDS       = "TIMEOUT_SECONDS"
	MAX_RETRIES           = "MAX_RETRIES"
	WAIT_INTERVAL_SECONDS = "WAIT_INTERVAL_SECONDS"
	BACKOFF_MULTIPLIER    = "BACKOFF_MULTIPLIER"
)

type retryerParams struct {
	TimeoutSeconds      int
	MaxRetries          int
	Multiplier          float64
	WaitIntervalSeconds int
}

type authedTransport struct {
	key     string
	wrapped http.RoundTripper
}

type Teams []struct {
	Name        string
	Description string
}

// Type aliases for genqlient stuff
type OrganizationTeams []getTeamsOrganizationTeamsTeamConnectionEdgesTeamEdge
type RateLimit getRateLimitResponse

func (t *authedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "bearer "+t.key)
	return t.wrapped.RoundTrip(req)
}

func parseEnv() *retryerParams {
	timeoutSeconds := 10
	maxRetries := 3
	multiplier := 2.0
	waitIntervalSeconds := 10

	if os.Getenv(TIMEOUT_SECONDS) != "" {
		TimeoutSeconds, err := strconv.Atoi(os.Getenv(TIMEOUT_SECONDS))
		if err != nil {
			panic(fmt.Errorf("can't parse %s: %w", TIMEOUT_SECONDS, err))
		}
		timeoutSeconds = TimeoutSeconds
	}
	if os.Getenv(MAX_RETRIES) != "" {
		MaxRetries, err := strconv.Atoi(os.Getenv(MAX_RETRIES))
		if err != nil {
			panic(fmt.Errorf("can't parse %s: %w", MAX_RETRIES, err))
		}
		maxRetries = MaxRetries
	}
	if os.Getenv(WAIT_INTERVAL_SECONDS) != "" {
		WaitIntervalSeconds, err := strconv.Atoi(os.Getenv(WAIT_INTERVAL_SECONDS))
		if err != nil {
			panic(fmt.Errorf("can't parse %s: %w", WAIT_INTERVAL_SECONDS, err))
		}
		waitIntervalSeconds = WaitIntervalSeconds
	}
	if os.Getenv(BACKOFF_MULTIPLIER) != "" {
		Multiplier, err := strconv.ParseFloat(os.Getenv(BACKOFF_MULTIPLIER), 64)
		if err != nil {
			panic(fmt.Errorf("can't parse %s: %w", BACKOFF_MULTIPLIER, err))
		}
		multiplier = Multiplier
	}

	// githubv4.Init("https://api.github.com/graphql", "", "")
	return &retryerParams{
		TimeoutSeconds:      timeoutSeconds,
		MaxRetries:          maxRetries,
		Multiplier:          multiplier,
		WaitIntervalSeconds: waitIntervalSeconds,
	}
}

func NewGHv4Client() *GitHubV4Client {
	parms := parseEnv()
	retryLogger := log.With().Str("component", "retryablehttp").Logger()
	gLog.SetOutput(retryLogger)

	key := os.Getenv("GITHUB_TOKEN")

	client := retryablehttp.NewClient()
	client.HTTPClient.Timeout = time.Second * time.Duration(parms.TimeoutSeconds)
	client.Logger = gLog.Default()
	client.RetryMax = parms.MaxRetries
	client.RetryWaitMax = time.Minute * time.Duration(parms.WaitIntervalSeconds) * time.Duration(parms.Multiplier)
	client.RetryWaitMin = time.Second * time.Duration(parms.WaitIntervalSeconds)
	client.HTTPClient.Transport = &authedTransport{
		key:     key,
		wrapped: http.DefaultTransport,
	}

	graphqlClient := graphql.NewClient("https://api.github.com/graphql", client.StandardClient())

	return &GitHubV4Client{
		baseClient:  client.StandardClient(),
		retryClient: client,
		client:      graphqlClient,
	}
}

func (c *GitHubV4Client) GetTeams() (OrganizationTeams, error) {
	orgTeams := []getTeamsOrganizationTeamsTeamConnectionEdgesTeamEdge{}
	initResp, err := getTeams(c.Context, c.client, TeamOrder{
		Direction: OrderDirectionDesc,
		Field:     TeamOrderFieldName,
	}, 100, "")
	if err != nil {
		panic(err)
	}
	log.Info().Interface("teams", initResp).Msg("initial teams")
	hasNextPage := initResp.Organization.Teams.PageInfo.HasNextPage
	cursor := initResp.Organization.Teams.PageInfo.EndCursor
	for hasNextPage {
		log.Debug().Str("cursor", cursor).Msg("has next page")
		resp, err := getTeams(c.Context, c.client, TeamOrder{
			Direction: OrderDirectionDesc,
			Field:     TeamOrderFieldName,
		}, 100, cursor)
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
	resp, err := getRateLimit(c.Context, c.client)
	if err != nil {
		return RateLimit{}, fmt.Errorf("can't get ratelimit: %w", err)
	}
	log.Debug().Interface("rate-limit", resp)
	return RateLimit(*resp), nil
}

func (c *GitHubV4Client) UpdateLabels(repo string, labels []Label) error {
	respositoryIssueLabe
	for _, label := range labels {
		_, err := updateLabel(c.Context, c.client, repo, label)
		if err != nil {
			return fmt.Errorf("can't update label: %w", err)
		}
	}
}
