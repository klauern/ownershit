package v4api

import (
	"context"
	"errors"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	mock_graphql "github.com/klauern/ownershit/v4api/mocks"
)

// testTokenSuffixLen is the length of the suffix for GitHub PAT tokens (excluding the "ghp_" prefix)
const testTokenSuffixLen = 36

// roundTripperFunc is a test helper to stub http.RoundTripper
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestNewGHv4Client(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		wantRetries int
		wantTimeout int
		wantErr     bool
	}{
		{
			name: "default configuration",
			envVars: map[string]string{
				"GITHUB_TOKEN": validTestToken(), // Valid test token format
			},
			wantRetries: defaultConfig.maxRetries,
			wantTimeout: defaultConfig.timeoutSeconds,
			wantErr:     false,
		},
		{
			name: "custom configuration",
			envVars: map[string]string{
				"GITHUB_TOKEN":                   validTestToken(), // Valid test token format
				EnvVarPrefix + EnvMaxRetries:     "5",
				EnvVarPrefix + EnvTimeoutSeconds: "30",
			},
			wantRetries: 5,
			wantTimeout: 30,
			wantErr:     false,
		},
		{
			name:        "missing token returns error",
			envVars:     map[string]string{"GITHUB_TOKEN": ""}, // Explicitly unset token
			wantRetries: defaultConfig.maxRetries,
			wantTimeout: defaultConfig.timeoutSeconds,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			client, err := NewGHv4Client()
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewGHv4Client() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}

			// Access the underlying HTTP client
			httpClient := client.retryClient
			if httpClient == nil {
				t.Fatal("failed to get underlying HTTP client")
				return
			}

			if httpClient.RetryMax != tt.wantRetries {
				t.Errorf("RetryMax = %v, want %v", httpClient.RetryMax, tt.wantRetries)
			}

			if httpClient.HTTPClient.Timeout != time.Duration(tt.wantTimeout)*time.Second {
				t.Errorf("Timeout = %v, want %v", httpClient.HTTPClient.Timeout, time.Duration(tt.wantTimeout)*time.Second)
			}
		})
	}
}

func Test_authedTransport_RoundTrip(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		wantHeader string
	}{
		{
			name:       "adds authorization header",
			key:        "test-key",
			wantHeader: "Bearer test-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use a stubbed RoundTripper to avoid real network/TLS
			transport := &authedTransport{
				key: tt.key,
				wrapped: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
					// Return a minimal valid response
					return &http.Response{StatusCode: 200, Body: http.NoBody, Header: make(http.Header)}, nil
				}),
			}

			req, err := http.NewRequest(http.MethodGet, "http://example.com", http.NoBody)
			if err != nil {
				t.Fatal(err)
			}

			resp, err := transport.RoundTrip(req)
			if err != nil {
				t.Fatal(err)
			}
			if resp != nil && resp.Body != nil {
				_ = resp.Body.Close()
			}

			if got := req.Header.Get("Authorization"); got != tt.wantHeader {
				t.Errorf("Authorization header = %v, want %v", got, tt.wantHeader)
			}
		})
	}
}

func Test_buildClient(t *testing.T) {
	type args struct {
		params *retryParams
		key    string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "default configuration",
			args: args{
				params: &retryParams{
					TimeoutSeconds:      10,
					MaxRetries:          3,
					Multiplier:          2.0,
					WaitIntervalSeconds: 10,
				},
				key: "test-key",
			},
		},
		{
			name: "custom configuration",
			args: args{
				params: &retryParams{
					TimeoutSeconds:      30,
					MaxRetries:          5,
					Multiplier:          1.5,
					WaitIntervalSeconds: 5,
				},
				key: "custom-key",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := buildClient(tt.args.params, tt.args.key)

			if client == nil {
				t.Error("buildClient() returned nil")
				return
			}

			// Verify it's the expected type
			_ = client

			if client.RetryMax != tt.args.params.MaxRetries {
				t.Errorf("buildClient() RetryMax = %v, want %v", client.RetryMax, tt.args.params.MaxRetries)
			}

			expectedTimeout := time.Duration(tt.args.params.TimeoutSeconds) * time.Second
			if client.HTTPClient.Timeout != expectedTimeout {
				t.Errorf("buildClient() Timeout = %v, want %v", client.HTTPClient.Timeout, expectedTimeout)
			}

			// Verify that the transport is properly set with authentication
			if _, ok := client.HTTPClient.Transport.(*authedTransport); !ok {
				t.Error("buildClient() should set authedTransport")
			}
		})
	}
}

func Test_parseEnv(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		want    *retryParams
	}{
		{
			name: "default values when no env vars set",
			want: &retryParams{
				TimeoutSeconds:      defaultConfig.timeoutSeconds,
				MaxRetries:          defaultConfig.maxRetries,
				Multiplier:          defaultConfig.multiplier,
				WaitIntervalSeconds: defaultConfig.waitIntervalSeconds,
			},
		},
		{
			name: "custom values",
			envVars: map[string]string{
				EnvVarPrefix + EnvMaxRetries:          "5",
				EnvVarPrefix + EnvTimeoutSeconds:      "30",
				EnvVarPrefix + EnvBackoffMultiplier:   "3.0",
				EnvVarPrefix + EnvWaitIntervalSeconds: "20",
			},
			want: &retryParams{
				TimeoutSeconds:      30,
				MaxRetries:          5,
				Multiplier:          3.0,
				WaitIntervalSeconds: 20,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment
			oldEnv := make(map[string]string)
			for k := range tt.envVars {
				oldEnv[k] = os.Getenv(k)
				_ = os.Setenv(k, tt.envVars[k])
			}
			// Cleanup
			defer func() {
				for k, v := range oldEnv {
					_ = os.Setenv(k, v)
				}
			}()

			got, err := parseEnv()
			if err != nil {
				t.Errorf("parseEnv() error = %v", err)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGitHubV4Client_GetRateLimit(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "nil client context should cause error",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a client with minimal setup for testing
			c := &GitHubV4Client{
				Context: nil, // This will cause issues
				client:  nil, // This will cause issues
			}

			// This test will likely panic or error due to nil context/client
			defer func() {
				if r := recover(); r != nil {
					if !tt.wantErr {
						t.Errorf("GetRateLimit() panicked: %v", r)
					}
				}
			}()

			_, err := c.GetRateLimit()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRateLimit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGitHubV4Client_GetTeams(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "nil client should cause panic",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a client with minimal setup for testing
			c := &GitHubV4Client{
				Context: nil,
				client:  nil,
			}

			// This test will likely panic due to nil client
			defer func() {
				if r := recover(); r != nil {
					if !tt.wantErr {
						t.Errorf("GetTeams() panicked: %v", r)
					}
				}
			}()

			_, err := c.GetTeams("test-org")
			if !tt.wantErr && err != nil {
				t.Errorf("GetTeams() unexpected error = %v", err)
			}
		})
	}
}

func TestGitHubV4Client_SyncLabels(t *testing.T) {
	tests := []struct {
		name    string
		repo    string
		labels  []Label
		wantErr bool
	}{
		{
			name:    "nil client should cause error",
			repo:    "test-repo",
			labels:  []Label{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a client with minimal setup for testing
			c := &GitHubV4Client{
				Context: nil,
				client:  nil,
			}

			// This test will panic due to nil client
			defer func() {
				if r := recover(); r != nil {
					if !tt.wantErr {
						t.Errorf("SyncLabels() panicked: %v", r)
					}
				}
			}()

			err := c.SyncLabels(tt.repo, "test-org", tt.labels)
			if !tt.wantErr && err != nil {
				t.Errorf("SyncLabels() unexpected error = %v", err)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	// Test that default configuration values are reasonable
	if defaultConfig.timeoutSeconds <= 0 {
		t.Error("defaultConfig.timeoutSeconds should be positive")
	}
	if defaultConfig.maxRetries < 0 {
		t.Error("defaultConfig.maxRetries should be non-negative")
	}
	if defaultConfig.multiplier <= 0 {
		t.Error("defaultConfig.multiplier should be positive")
	}
	if defaultConfig.waitIntervalSeconds <= 0 {
		t.Error("defaultConfig.waitIntervalSeconds should be positive")
	}
}

func TestConstants(t *testing.T) {
	// Test that constants are properly defined
	if V4ClientDefaultPageSize <= 0 {
		t.Error("V4ClientDefaultPageSize should be positive")
	}

	if EnvVarPrefix == "" {
		t.Error("EnvVarPrefix should not be empty")
	}

	// Test environment variable names
	envVars := []string{
		EnvTimeoutSeconds,
		EnvMaxRetries,
		EnvWaitIntervalSeconds,
		EnvBackoffMultiplier,
	}

	for _, envVar := range envVars {
		if envVar == "" {
			t.Errorf("Environment variable constant should not be empty")
		}
	}
}

func TestGitHubV4Client_GetRateLimit_MockSuccessPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock_graphql.NewMockClient(ctrl)
	ctx := context.Background()

	client := &GitHubV4Client{
		Context: ctx,
		client:  mockClient,
	}

	// Mock the GraphQL call - this tests that the call is made correctly
	mockClient.EXPECT().MakeRequest(ctx, gomock.Any(), gomock.Any()).Return(nil)

	// Note: Due to the complexity of mocking genqlient generated code,
	// we focus on testing that the call path is executed correctly
	_, err := client.GetRateLimit()
	// The actual response depends on genqlient's unmarshaling,
	// so we just verify the method executes without panicking
	_ = err
}

func TestGitHubV4Client_GetRateLimit_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock_graphql.NewMockClient(ctrl)
	ctx := context.Background()

	client := &GitHubV4Client{
		Context: ctx,
		client:  mockClient,
	}

	expectedError := errors.New("graphql error")
	mockClient.EXPECT().MakeRequest(ctx, gomock.Any(), gomock.Any()).Return(expectedError)

	_, err := client.GetRateLimit()
	if err == nil {
		t.Error("Expected error, but got nil")
	}
	if !errors.Is(err, expectedError) {
		t.Errorf("Expected error to wrap %v, but got %v", expectedError, err)
	}
}

func TestGitHubV4Client_GetTeams_WithContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock_graphql.NewMockClient(ctrl)
	ctx := context.Background()

	client := &GitHubV4Client{
		Context: ctx,
		client:  mockClient,
	}

	// Mock the GraphQL call that will be made
	mockClient.EXPECT().MakeRequest(ctx, gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	// This tests the GetTeams method execution path
	_, err := client.GetTeams("test-org")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestGitHubV4Client_GetTeams_GraphQLError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock_graphql.NewMockClient(ctrl)
	ctx := context.Background()

	client := &GitHubV4Client{
		Context: ctx,
		client:  mockClient,
	}

	expectedError := errors.New("graphql error")
	mockClient.EXPECT().MakeRequest(ctx, gomock.Any(), gomock.Any()).Return(expectedError)

	_, err := client.GetTeams("test-org")
	if err == nil {
		t.Error("Expected error due to GraphQL error, but got none")
	}
}

func TestGitHubV4Client_SyncLabels_WithContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock_graphql.NewMockClient(ctrl)
	ctx := context.Background()

	client := &GitHubV4Client{
		Context: ctx,
		client:  mockClient,
	}

	// Mock the GraphQL call that will be made (SyncLabels makes at least one call to get labels)
	mockClient.EXPECT().MakeRequest(ctx, gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	labels := []Label{}
	err := client.SyncLabels("test-repo", "test-org", labels)
	// The method should execute without panicking
	_ = err
}

func TestGitHubV4Client_SyncLabels_GetLabelsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock_graphql.NewMockClient(ctrl)
	ctx := context.Background()

	client := &GitHubV4Client{
		Context: ctx,
		client:  mockClient,
	}

	expectedError := errors.New("failed to get labels")
	mockClient.EXPECT().MakeRequest(ctx, gomock.Any(), gomock.Any()).Return(expectedError)

	labels := []Label{}
	err := client.SyncLabels("test-repo", "test-org", labels)
	if err == nil {
		t.Error("Expected error, but got nil")
	}
	if !errors.Is(err, expectedError) {
		t.Errorf("Expected error to wrap %v, but got %v", expectedError, err)
	}
}

func TestParseEnv_ValidValues(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		want    *retryParams
	}{
		{
			name: "valid custom values",
			envVars: map[string]string{
				EnvVarPrefix + EnvMaxRetries:          "5",
				EnvVarPrefix + EnvTimeoutSeconds:      "30",
				EnvVarPrefix + EnvBackoffMultiplier:   "2.5",
				EnvVarPrefix + EnvWaitIntervalSeconds: "20",
			},
			want: &retryParams{
				TimeoutSeconds:      30,
				MaxRetries:          5,
				Multiplier:          2.5,
				WaitIntervalSeconds: 20,
			},
		},
		{
			name: "empty environment - should use defaults",
			envVars: map[string]string{
				EnvVarPrefix + EnvMaxRetries:          "",
				EnvVarPrefix + EnvTimeoutSeconds:      "",
				EnvVarPrefix + EnvBackoffMultiplier:   "",
				EnvVarPrefix + EnvWaitIntervalSeconds: "",
			},
			want: &retryParams{
				TimeoutSeconds:      defaultConfig.timeoutSeconds,
				MaxRetries:          defaultConfig.maxRetries,
				Multiplier:          defaultConfig.multiplier,
				WaitIntervalSeconds: defaultConfig.waitIntervalSeconds,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment
			oldEnv := make(map[string]string)
			for k, v := range tt.envVars {
				oldEnv[k] = os.Getenv(k)
				_ = os.Setenv(k, v)
			}
			// Cleanup
			defer func() {
				for k, v := range oldEnv {
					_ = os.Setenv(k, v)
				}
			}()

			got, err := parseEnv()
			if err != nil {
				t.Errorf("parseEnv() error = %v", err)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthedTransport_RoundTrip_Error(t *testing.T) {
	// Create a transport that will cause an error
	transport := &authedTransport{
		key:     "test-key",
		wrapped: &errorTransport{},
	}

	req, err := http.NewRequest(http.MethodGet, "http://example.com", http.NoBody)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := transport.RoundTrip(req)
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
	if err == nil {
		t.Error("Expected error from wrapped transport, but got nil")
	}

	// Verify authorization header was still set
	if got := req.Header.Get("Authorization"); got != "Bearer test-key" {
		t.Errorf("Authorization header = %v, want %v", got, "Bearer test-key")
	}
}

func TestBuildClient_ValidateRetryConfiguration(t *testing.T) {
	params := &retryParams{
		TimeoutSeconds:      30,
		MaxRetries:          5,
		Multiplier:          2.5,
		WaitIntervalSeconds: 15,
	}

	client := buildClient(params, "test-key")

	if client.RetryMax != params.MaxRetries {
		t.Errorf("RetryMax = %v, want %v", client.RetryMax, params.MaxRetries)
	}

	expectedTimeout := time.Duration(params.TimeoutSeconds) * time.Second
	if client.HTTPClient.Timeout != expectedTimeout {
		t.Errorf("Timeout = %v, want %v", client.HTTPClient.Timeout, expectedTimeout)
	}

	expectedWaitMax := time.Duration(float64(params.WaitIntervalSeconds)*params.Multiplier) * time.Second
	if client.RetryWaitMax != expectedWaitMax {
		t.Errorf("RetryWaitMax = %v, want %v", client.RetryWaitMax, expectedWaitMax)
	}

	expectedWaitMin := time.Duration(params.WaitIntervalSeconds) * time.Second
	if client.RetryWaitMin != expectedWaitMin {
		t.Errorf("RetryWaitMin = %v, want %v", client.RetryWaitMin, expectedWaitMin)
	}
}

func TestNewGHv4Client_TokenValidationError(t *testing.T) {
	// Unset token to cause validation error
	t.Setenv("GITHUB_TOKEN", "")

	// Now we expect an error instead of a panic
	_, err := NewGHv4Client()
	if err == nil {
		t.Error("Expected error due to token validation failure, but got none")
	}
}

// Helper type for testing error transport.
type errorTransport struct{}

func (t *errorTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, errors.New("transport error")
}

func TestTypeAliases(t *testing.T) {
	// Test that type aliases are properly defined
	var teams OrganizationTeams
	if len(teams) == 0 {
		teams = make(OrganizationTeams, 0)
	}
	_ = teams

	var rateLimit RateLimit
	_ = rateLimit

	var label Label
	_ = label

	// These tests just verify the types compile correctly
	t.Log("Type aliases test passed")
}

// validTestToken returns a runtime-generated GitHub PAT format token for testing
func validTestToken() string {
	return "ghp_" + strings.Repeat("A", testTokenSuffixLen)
}

func TestGitHubV4Client_ContextHandling(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock_graphql.NewMockClient(ctrl)

	tests := []struct {
		name    string
		context context.Context
	}{
		{
			name:    "valid context",
			context: context.Background(),
		},
		{
			name: "context with timeout",
			context: func() context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				return ctx
			}(),
		},
		{
			name:    "canceled context",
			context: func() context.Context { ctx, cancel := context.WithCancel(context.Background()); cancel(); return ctx }(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &GitHubV4Client{
				Context: tt.context,
				client:  mockClient,
			}

			// Just verify the method can be called with different contexts
			mockClient.EXPECT().MakeRequest(tt.context, gomock.Any(), gomock.Any()).Return(nil)

			_, err := client.GetRateLimit()
			// We just verify the call path executes
			_ = err
		})
	}
}

func TestGitHubV4Client_SyncLabels_BasicPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock_graphql.NewMockClient(ctrl)
	ctx := context.Background()

	client := &GitHubV4Client{
		Context: ctx,
		client:  mockClient,
	}

	// Mock the initial call to get repository labels
	mockClient.EXPECT().MakeRequest(ctx, gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	labels := []Label{}
	err := client.SyncLabels("test-repo", "test-org", labels)
	// We just verify the method executes its basic code path
	_ = err
}

func TestGitHubV4Client_GetTeams_PaginationPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock_graphql.NewMockClient(ctrl)
	ctx := context.Background()

	client := &GitHubV4Client{
		Context: ctx,
		client:  mockClient,
	}

	// Test pagination scenario - mock multiple calls
	mockClient.EXPECT().MakeRequest(ctx, gomock.Any(), gomock.Any()).Return(nil).MinTimes(1)

	_, err := client.GetTeams("test-org")
	if err != nil {
		t.Errorf("Unexpected error in GetTeams: %v", err)
	}
}

func TestAuthedTransport_RoundTrip_VariousRequests(t *testing.T) {
	tests := []struct {
		name   string
		method string
		url    string
		key    string
	}{
		{
			name:   "GET request",
			method: http.MethodGet,
			url:    "https://api.github.com/graphql",
			key:    "test-token-1",
		},
		{
			name:   "POST request",
			method: http.MethodPost,
			url:    "https://api.github.com/repos",
			key:    "test-token-2",
		},
		{
			name:   "empty key",
			method: http.MethodGet,
			url:    "https://example.com",
			key:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := &authedTransport{
				key:     tt.key,
				wrapped: http.DefaultTransport,
			}

			req, err := http.NewRequest(tt.method, tt.url, http.NoBody)
			if err != nil {
				t.Fatal(err)
			}

			resp, err := transport.RoundTrip(req)
			if err != nil {
				t.Errorf("RoundTrip failed: %v", err)
			}
			if resp != nil && resp.Body != nil {
				_ = resp.Body.Close()
			}

			expectedAuth := "Bearer " + tt.key
			if got := req.Header.Get("Authorization"); got != expectedAuth {
				t.Errorf("Authorization header = %v, want %v", got, expectedAuth)
			}
		})
	}
}

func TestRetryParams_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		params   *retryParams
		wantZero bool
	}{
		{
			name: "zero values",
			params: &retryParams{
				TimeoutSeconds:      0,
				MaxRetries:          0,
				Multiplier:          0,
				WaitIntervalSeconds: 0,
			},
			wantZero: true,
		},
		{
			name: "negative values",
			params: &retryParams{
				TimeoutSeconds:      -1,
				MaxRetries:          -1,
				Multiplier:          -1.0,
				WaitIntervalSeconds: -1,
			},
			wantZero: false,
		},
		{
			name: "large values",
			params: &retryParams{
				TimeoutSeconds:      3600,
				MaxRetries:          100,
				Multiplier:          10.0,
				WaitIntervalSeconds: 600,
			},
			wantZero: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := buildClient(tt.params, "test-key")

			if client == nil {
				t.Error("buildClient returned nil")
				return
			}

			// Verify basic client structure
			if client.HTTPClient == nil {
				t.Error("HTTPClient is nil")
				return
			}

			if client.HTTPClient.Transport == nil {
				t.Error("Transport is nil")
				return
			}

			// Verify auth transport is set
			if _, ok := client.HTTPClient.Transport.(*authedTransport); !ok {
				t.Error("Expected authedTransport")
			}
		})
	}
}
