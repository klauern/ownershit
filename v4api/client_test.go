package v4api

import (
	"net/http"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	mockgraphql "github.com/klauern/ownershit/v4api/mocks"
	"go.uber.org/mock/gomock"
)

type testMocks struct {
	ctrl        *gomock.Controller
	graphQLMock *mockgraphql.MockClient
}

func setupMocks(t *testing.T) *testMocks {
	ctrl := gomock.NewController(t)
	graphQLClient := mockgraphql.NewMockClient(ctrl)

	return &testMocks{
		ctrl:        ctrl,
		graphQLMock: graphQLClient,
	}
}

func TestNewGHv4Client(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		wantRetries int
		wantTimeout int
	}{
		{
			name:        "default configuration",
			wantRetries: defaultConfig.maxRetries,
			wantTimeout: defaultConfig.timeoutSeconds,
		},
		{
			name: "custom configuration",
			envVars: map[string]string{
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
			for k, v := range tt.envVars {
				oldVal := os.Getenv(k)
				os.Setenv(k, v)
				defer os.Setenv(k, oldVal)
			}

			client := NewGHv4Client()

			// Access the underlying HTTP client
			httpClient := client.retryClient
			if httpClient == nil {
				t.Fatal("failed to get underlying HTTP client")
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

			req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
			if err != nil {
				t.Fatal(err)
			}

			_, err = transport.RoundTrip(req)
			if err != nil {
				t.Fatal(err)
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
		want *retryablehttp.Client
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildClient(tt.args.params, tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildClient() = %v, want %v", got, tt.want)
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
