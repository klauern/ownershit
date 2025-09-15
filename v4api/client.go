// Package v4api provides GitHub GraphQL API v4 client functionality.
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

	"github.com/klauern/ownershit"
)

//go:generate mockgen -destination mocks/client_mocks.go github.com/Khan/genqlient/graphql Client

// GitHubV4Client provides a GraphQL client for GitHub API v4 operations.
type GitHubV4Client struct {
	baseClient  *http.Client
	retryClient *retryablehttp.Client
	client      graphql.Client
	Context     context.Context
}

// EnvVarPrefix is the prefix used for environment variable names.
const EnvVarPrefix = "OWNERSHIT_"

var defaultConfig = struct {
	timeoutSeconds      int
	maxRetries          int
	multiplier          float64
	waitIntervalSeconds int
}{
	timeoutSeconds:      10,
	maxRetries:          3,
	multiplier:          2.0,
	waitIntervalSeconds: 10,
}

const (
	// EnvTimeoutSeconds is the environment variable for timeout configuration.
	EnvTimeoutSeconds = "TIMEOUT_SECONDS"
	// EnvMaxRetries is the environment variable for retry configuration.
	EnvMaxRetries = "MAX_RETRIES"
	// EnvWaitIntervalSeconds is the environment variable for wait interval configuration.
	EnvWaitIntervalSeconds = "WAIT_INTERVAL_SECONDS"
	// EnvBackoffMultiplier is the environment variable for backoff multiplier configuration.
	EnvBackoffMultiplier = "BACKOFF_MULTIPLIER"
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

// Teams represents a collection of GitHub teams.
type Teams []struct {
	Name        string
	Description string
}

type (
	// OrganizationTeams represents teams in a GitHub organization.
	OrganizationTeams []GetTeamsOrganizationTeamsTeamConnectionEdgesTeamEdge
	// RateLimit represents GitHub API rate limit information.
	RateLimit GetRateLimitResponse
	// Label represents a GitHub repository label.
	Label GetRepositoryIssueLabelsRepositoryLabelsLabelConnectionEdgesLabelEdgeNodeLabel
)

func (t *authedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.key != "" {
		req.Header.Set("Authorization", "Bearer "+t.key)
	}
	return t.wrapped.RoundTrip(req)
}

func parseEnv() (*retryParams, error) {
	params := &retryParams{
		TimeoutSeconds:      defaultConfig.timeoutSeconds,
		MaxRetries:          defaultConfig.maxRetries,
		Multiplier:          defaultConfig.multiplier,
		WaitIntervalSeconds: defaultConfig.waitIntervalSeconds,
	}

	parseEnvInt := func(env string, target *int) error {
		if val := os.Getenv(EnvVarPrefix + env); val != "" {
			if parsed, err := strconv.Atoi(val); err == nil {
				*target = parsed
			} else {
				log.Warn().Err(err).Str("env", env).Msg("invalid env; using default")
				return err
			}
		}
		return nil
	}

	if err := parseEnvInt(EnvTimeoutSeconds, &params.TimeoutSeconds); err != nil {
		return nil, err
	}
	if err := parseEnvInt(EnvMaxRetries, &params.MaxRetries); err != nil {
		return nil, err
	}
	if err := parseEnvInt(EnvWaitIntervalSeconds, &params.WaitIntervalSeconds); err != nil {
		return nil, err
	}

	if val := os.Getenv(EnvVarPrefix + EnvBackoffMultiplier); val != "" {
		if parsed, err := strconv.ParseFloat(val, 64); err == nil {
			params.Multiplier = parsed
		} else {
			log.Warn().Err(err).Str("env", EnvBackoffMultiplier).Msg("invalid env; using default")
			return nil, err
		}
	}

	return params, nil
}

// NewGHv4Client constructs an authenticated GitHub GraphQL v4 client with
// retry/backoff configuration derived from environment variables prefixed with
// OWNERSHIT_. The GitHub token is validated before creating the client.
func NewGHv4Client() (*GitHubV4Client, error) {
	params, err := parseEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to parse environment configuration: %w", err)
	}
	// Note: stdlib log expects an io.Writer; keep default output. RetryableHTTP logging
	// is wired to the stdlib logger in buildClient().
	key, err := ownershit.GetValidatedGitHubToken()
	if err != nil {
		return nil, fmt.Errorf("GitHub token validation failed: %w", err)
	}

	client := buildClient(params, key)

	graphqlClient := graphql.NewClient("https://api.github.com/graphql", client.StandardClient())

	return &GitHubV4Client{
		baseClient:  client.StandardClient(),
		retryClient: client,
		client:      graphqlClient,
		Context:     context.Background(),
	}, nil
}

// buildClient sets up the retry functionality and attaches authentication to the client.
func buildClient(params *retryParams, key string) *retryablehttp.Client {
	client := retryablehttp.NewClient()
	client.HTTPClient.Timeout = time.Duration(params.TimeoutSeconds) * time.Second
	client.Logger = globalLog.Default()
	client.RetryMax = params.MaxRetries
	client.RetryWaitMax = time.Duration(float64(params.WaitIntervalSeconds)*params.Multiplier) * time.Second
	client.RetryWaitMin = time.Duration(params.WaitIntervalSeconds) * time.Second
	client.HTTPClient.Transport = &authedTransport{
		key:     key,
		wrapped: http.DefaultTransport,
	}
	return client
}
