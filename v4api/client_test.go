package v4api

import (
	"net/http"
	"os"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hashicorp/go-retryablehttp"
	mockgraphql "github.com/klauern/ownershit/v4api/mocks"
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
		name string
		want *GitHubV4Client
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewGHv4Client(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewGHv4Client() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_authedTransport_RoundTrip(t1 *testing.T) {
	type fields struct {
		key     string
		wrapped http.RoundTripper
	}
	type args struct {
		req *http.Request
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *http.Response
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &authedTransport{
				key:     tt.fields.key,
				wrapped: tt.fields.wrapped,
			}
			got, err := t.RoundTrip(tt.args.req)
			if (err != nil) != tt.wantErr {
				t1.Errorf("RoundTrip() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t1.Errorf("RoundTrip() got = %v, want %v", got, tt.want)
			}
			got.Body.Close()
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
			name: "partial evaluation 1",
			envVars: map[string]string{
				"OWNERSHIT_MAX_RETRIES": "10",
			},
			want: &retryParams{
				TimeoutSeconds:      ClientDefaultTimeoutSeconds,
				MaxRetries:          10,
				Multiplier:          ClientDefaultMultiplier,
				WaitIntervalSeconds: ClientDefaultWaitIntervalSeconds,
			},
		},
		{
			name: "partial evaluation 2",
			envVars: map[string]string{
				"OWNERSHIT_MAX_RETRIES":     "1",
				"OWNERSHIT_TIMEOUT_SECONDS": "12345",
			},
			want: &retryParams{
				TimeoutSeconds:      12345,
				MaxRetries:          1,
				Multiplier:          ClientDefaultMultiplier,
				WaitIntervalSeconds: ClientDefaultWaitIntervalSeconds,
			},
		},
		{
			name: "override all",
			envVars: map[string]string{
				"OWNERSHIT_MAX_RETRIES":           "1",
				"OWNERSHIT_TIMEOUT_SECONDS":       "12345",
				"OWNERSHIT_BACKOFF_MULTIPLIER":    "4546",
				"OWNERSHIT_WAIT_INTERVAL_SECONDS": "789",
			},
			want: &retryParams{
				MaxRetries:          1,
				TimeoutSeconds:      12345,
				Multiplier:          4546.0,
				WaitIntervalSeconds: 789,
			},
		},
	}
	for _, tt := range tests {
		oldEnviron := map[string]string{}
		for k, v := range tt.envVars {
			oldEnviron[k] = os.Getenv(k)
			err := os.Setenv(k, v)
			if err != nil {
				t.Errorf("unable to set environment variable %s: %v", k, err)
			}
		}
		t.Run(tt.name, func(t *testing.T) {
			if got := parseEnv(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseEnv() = %v, want %v", got, tt.want)
			}
		})
		for k, v := range oldEnviron {
			os.Setenv(k, v)
		}
	}
}
