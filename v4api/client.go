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
	EnvVarPrefix           = "OWNERSHIT_"
	EnvTimeoutSeconds      = "TIMEOUT_SECONDS"
	EnvMaxRetries          = "MAX_RETRIES"
	EnvWaitIntervalSeconds = "WAIT_INTERVAL_SECONDS"
	EnvBackoffMultiplier   = "BACKOFF_MULTIPLIER"

	ClientDefaultTimeoutSeconds      = 10
	ClientDefaultMaxRetries          = 3
	ClientDefaultMultiplier          = 2.0
	ClientDefaultWaitIntervalSeconds = 10
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
	timeoutSeconds := ClientDefaultTimeoutSeconds
	maxRetries := ClientDefaultMaxRetries
	multiplier := ClientDefaultMultiplier
	waitIntervalSeconds := ClientDefaultWaitIntervalSeconds

	if os.Getenv(EnvVarPrefix+EnvTimeoutSeconds) != "" {
		parsedTimeoutSeconds, err := strconv.Atoi(os.Getenv(EnvVarPrefix + EnvTimeoutSeconds))
		if err != nil {
			panic(fmt.Errorf("can't parse %s: %w", EnvTimeoutSeconds, err))
		}
		timeoutSeconds = parsedTimeoutSeconds
	}
	if os.Getenv(EnvVarPrefix+EnvMaxRetries) != "" {
		parsedMaxRetries, err := strconv.Atoi(os.Getenv(EnvVarPrefix + EnvMaxRetries))
		if err != nil {
			panic(fmt.Errorf("can't parse %s: %w", EnvMaxRetries, err))
		}
		maxRetries = parsedMaxRetries
	}
	if os.Getenv(EnvVarPrefix+EnvWaitIntervalSeconds) != "" {
		parsedWaitIntervalSeconds, err := strconv.Atoi(os.Getenv(EnvVarPrefix + EnvWaitIntervalSeconds))
		if err != nil {
			panic(fmt.Errorf("can't parse %s: %w", EnvWaitIntervalSeconds, err))
		}
		waitIntervalSeconds = parsedWaitIntervalSeconds
	}
	if os.Getenv(EnvVarPrefix+EnvBackoffMultiplier) != "" {
		parsedMultiplier, err := strconv.ParseFloat(os.Getenv(EnvVarPrefix+EnvBackoffMultiplier), 64)
		if err != nil {
			panic(fmt.Errorf("can't parse %s: %w", EnvBackoffMultiplier, err))
		}
		multiplier = parsedMultiplier
	}

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
