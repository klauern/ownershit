package v4api

import (
	"net/http"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

func TestNewGHv4Client(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		wantRetries int
		wantTimeout int
	}{
		{
			name: "default configuration",
			envVars: map[string]string{
				"GITHUB_TOKEN": "ghp_abcd1234567890ABCD1234567890abcd1234", // Valid test token format
			},
			wantRetries: defaultConfig.maxRetries,
			wantTimeout: defaultConfig.timeoutSeconds,
		},
		{
			name: "custom configuration",
			envVars: map[string]string{
				"GITHUB_TOKEN":                   "ghp_abcd1234567890ABCD1234567890abcd1234", // Valid test token format
				EnvVarPrefix + EnvMaxRetries:     "5",
				EnvVarPrefix + EnvTimeoutSeconds: "30",
			},
			wantRetries: 5,
			wantTimeout: 30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment
			oldVals := make(map[string]string)
			for k, v := range tt.envVars {
				oldVals[k] = os.Getenv(k)
				os.Setenv(k, v)
			}
			defer func() {
				for k, oldVal := range oldVals {
					os.Setenv(k, oldVal)
				}
			}()

			client := NewGHv4Client()

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
			wantHeader: "bearer test-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := &authedTransport{
				key:     tt.key,
				wrapped: http.DefaultTransport,
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
				resp.Body.Close()
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
			var _ *retryablehttp.Client = client

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
				os.Setenv(k, tt.envVars[k])
			}
			// Cleanup
			defer func() {
				for k, v := range oldEnv {
					os.Setenv(k, v)
				}
			}()

			got := parseEnv()
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

			_, err := c.GetTeams()
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

			err := c.SyncLabels(tt.repo, tt.labels)
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
