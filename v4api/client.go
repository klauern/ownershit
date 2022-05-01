package v4api

import (
	"context"
	"fmt"
	globalLog "log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/rs/zerolog/log"
)

//go:generate mockgen -destination mocks/client_mocks.go github.com/Khan/genqlient/graphql Client

type GitHubV4Client struct {
	baseClient  *http.Client
	retryClient *retryablehttp.Client
	client      graphql.Client
	Context     context.Context
}

const (
	ENV_VAR_PREFIX        = "OWNERSHIT_"
	TIMEOUT_SECONDS       = "TIMEOUT_SECONDS"
	MAX_RETRIES           = "MAX_RETRIES"
	WAIT_INTERVAL_SECONDS = "WAIT_INTERVAL_SECONDS"
	BACKOFF_MULTIPLIER    = "BACKOFF_MULTIPLIER"
)

type retryParams struct {
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

// various type aliases around genqlient
type (
	OrganizationTeams []GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdge
	RateLimit         GetRateLimitResponse
	Label             GetRepositoryIssueLabelsRepositoryLabelsLabelConnectionEdgesLabelEdgeNodeLabel
)

func (t *authedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "bearer "+t.key)
	return t.wrapped.RoundTrip(req)
}

func parseEnv() *retryParams {
	// default values
	timeoutSeconds := 10
	maxRetries := 3
	multiplier := 2.0
	waitIntervalSeconds := 10

	if os.Getenv(ENV_VAR_PREFIX+TIMEOUT_SECONDS) != "" {
		TimeoutSeconds, err := strconv.Atoi(os.Getenv(ENV_VAR_PREFIX + TIMEOUT_SECONDS))
		if err != nil {
			panic(fmt.Errorf("can't parse %s: %w", TIMEOUT_SECONDS, err))
		}
		timeoutSeconds = TimeoutSeconds
	}
	if os.Getenv(ENV_VAR_PREFIX+MAX_RETRIES) != "" {
		MaxRetries, err := strconv.Atoi(os.Getenv(ENV_VAR_PREFIX + MAX_RETRIES))
		if err != nil {
			panic(fmt.Errorf("can't parse %s: %w", MAX_RETRIES, err))
		}
		maxRetries = MaxRetries
	}
	if os.Getenv(ENV_VAR_PREFIX+WAIT_INTERVAL_SECONDS) != "" {
		WaitIntervalSeconds, err := strconv.Atoi(os.Getenv(ENV_VAR_PREFIX + WAIT_INTERVAL_SECONDS))
		if err != nil {
			panic(fmt.Errorf("can't parse %s: %w", WAIT_INTERVAL_SECONDS, err))
		}
		waitIntervalSeconds = WaitIntervalSeconds
	}
	if os.Getenv(ENV_VAR_PREFIX+BACKOFF_MULTIPLIER) != "" {
		Multiplier, err := strconv.ParseFloat(os.Getenv(ENV_VAR_PREFIX+BACKOFF_MULTIPLIER), 64)
		if err != nil {
			panic(fmt.Errorf("can't parse %s: %w", BACKOFF_MULTIPLIER, err))
		}
		multiplier = Multiplier
	}

	// githubv4.Init("https://api.github.com/graphql", "", "")
	return &retryParams{
		TimeoutSeconds:      timeoutSeconds,
		MaxRetries:          maxRetries,
		Multiplier:          multiplier,
		WaitIntervalSeconds: waitIntervalSeconds,
	}
}

func NewGHv4Client() *GitHubV4Client {
	params := parseEnv()
	retryLogger := log.With().Str("component", "retryablehttp").Logger()
	globalLog.SetOutput(retryLogger)

	key := os.Getenv("GITHUB_TOKEN")

	client := buildClient(params, key)

	graphqlClient := graphql.NewClient("https://api.github.com/graphql", client.StandardClient())

	return &GitHubV4Client{
		baseClient:  client.StandardClient(),
		retryClient: client,
		client:      graphqlClient,
	}
}

// buildClient sets up the retry functionality and attaches authentication to the client
func buildClient(params *retryParams, key string) *retryablehttp.Client {
	client := retryablehttp.NewClient()
	client.HTTPClient.Timeout = time.Second * time.Duration(params.TimeoutSeconds)
	client.Logger = globalLog.Default()
	client.RetryMax = params.MaxRetries
	client.RetryWaitMax = time.Minute * time.Duration(params.WaitIntervalSeconds) * time.Duration(params.Multiplier)
	client.RetryWaitMin = time.Second * time.Duration(params.WaitIntervalSeconds)
	client.HTTPClient.Transport = &authedTransport{
		key:     key,
		wrapped: http.DefaultTransport,
	}
	return client
}
