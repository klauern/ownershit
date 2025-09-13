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

type GitHubV4Client struct {
	baseClient  *http.Client
	retryClient *retryablehttp.Client
	client      graphql.Client
	Context     context.Context
}

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
	EnvTimeoutSeconds      = "TIMEOUT_SECONDS"
	EnvMaxRetries          = "MAX_RETRIES"
	EnvWaitIntervalSeconds = "WAIT_INTERVAL_SECONDS"
	EnvBackoffMultiplier   = "BACKOFF_MULTIPLIER"
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

// Various type aliases around genqlient.
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
	params := &retryParams{
		TimeoutSeconds:      defaultConfig.timeoutSeconds,
		MaxRetries:          defaultConfig.maxRetries,
		Multiplier:          defaultConfig.multiplier,
		WaitIntervalSeconds: defaultConfig.waitIntervalSeconds,
	}

	parseEnvInt := func(env string, target *int) {
		if val := os.Getenv(EnvVarPrefix + env); val != "" {
			if parsed, err := strconv.Atoi(val); err == nil {
				*target = parsed
			} else {
				log.Fatal().Err(err).Str("env", env).Msg("failed to parse environment variable")
			}
		}
	}

	parseEnvInt(EnvTimeoutSeconds, &params.TimeoutSeconds)
	parseEnvInt(EnvMaxRetries, &params.MaxRetries)
	parseEnvInt(EnvWaitIntervalSeconds, &params.WaitIntervalSeconds)

	if val := os.Getenv(EnvVarPrefix + EnvBackoffMultiplier); val != "" {
		if parsed, err := strconv.ParseFloat(val, 64); err == nil {
			params.Multiplier = parsed
		} else {
			log.Fatal().Err(err).Str("env", EnvBackoffMultiplier).Msg("failed to parse environment variable")
		}
	}

	return params
}

func NewGHv4Client() (*GitHubV4Client, error) {
    params := parseEnv()
    retryLogger := log.With().Str("component", "retryablehttp").Logger()
    globalLog.SetOutput(retryLogger)

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
    }, nil
}

// buildClient sets up the retry functionality and attaches authentication to the client.
func buildClient(params *retryParams, key string) *retryablehttp.Client {
	client := retryablehttp.NewClient()
	client.HTTPClient.Timeout = time.Duration(params.TimeoutSeconds) * time.Second
	client.Logger = globalLog.Default()
	client.RetryMax = params.MaxRetries
	client.RetryWaitMax = time.Duration(params.WaitIntervalSeconds) * time.Duration(params.Multiplier) * time.Minute
	client.RetryWaitMin = time.Duration(params.WaitIntervalSeconds) * time.Second
	client.HTTPClient.Transport = &authedTransport{
		key:     key,
		wrapped: http.DefaultTransport,
	}
	return client
}
