package v4api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestBuildClient_ConfigurationTest(t *testing.T) {
	params := &retryParams{
		TimeoutSeconds:      30,
		MaxRetries:          5,
		Multiplier:          2.0,
		WaitIntervalSeconds: 10,
	}
	key := "test-key"

	client := buildClient(params, key)

	// Test all the configuration aspects
	if client.RetryMax != 5 {
		t.Errorf("Expected RetryMax 5, got %d", client.RetryMax)
	}

	expectedTimeout := 30 * time.Second
	if client.HTTPClient.Timeout != expectedTimeout {
		t.Errorf("Expected timeout %v, got %v", expectedTimeout, client.HTTPClient.Timeout)
	}

	expectedWaitMin := 10 * time.Second
	if client.RetryWaitMin != expectedWaitMin {
		t.Errorf("Expected RetryWaitMin %v, got %v", expectedWaitMin, client.RetryWaitMin)
	}

	expectedWaitMax := time.Duration(10) * time.Duration(2.0) * time.Minute
	if client.RetryWaitMax != expectedWaitMax {
		t.Errorf("Expected RetryWaitMax %v, got %v", expectedWaitMax, client.RetryWaitMax)
	}

	// Test the auth transport
	if transport, ok := client.HTTPClient.Transport.(*authedTransport); ok {
		if transport.key != key {
			t.Errorf("Expected key %s, got %s", key, transport.key)
		}
		if transport.wrapped != http.DefaultTransport {
			t.Error("Expected wrapped transport to be default")
		}
	} else {
		t.Error("Expected authedTransport")
	}
}

func TestAuthedTransport_RoundTrip_Coverage(t *testing.T) {
	transport := &authedTransport{
		key:     "test-token",
		wrapped: http.DefaultTransport,
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Expected Authorization 'Bearer test-token', got '%s'", got)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	req, _ := http.NewRequest("GET", srv.URL, http.NoBody)
	if _, err := transport.RoundTrip(req); err != nil {
		t.Fatalf("RoundTrip failed: %v", err)
	}
}

func TestParseEnv_CoverageTest(t *testing.T) {
	// Save original environment
	originalEnv := map[string]string{
		EnvVarPrefix + EnvTimeoutSeconds:      os.Getenv(EnvVarPrefix + EnvTimeoutSeconds),
		EnvVarPrefix + EnvMaxRetries:          os.Getenv(EnvVarPrefix + EnvMaxRetries),
		EnvVarPrefix + EnvWaitIntervalSeconds: os.Getenv(EnvVarPrefix + EnvWaitIntervalSeconds),
		EnvVarPrefix + EnvBackoffMultiplier:   os.Getenv(EnvVarPrefix + EnvBackoffMultiplier),
	}

	// Clean up after test
	defer func() {
		for k, v := range originalEnv {
			if v == "" {
				os.Unsetenv(k)
			} else {
				_ = os.Setenv(k, v)
			}
		}
	}()

	// Test with no environment variables set
	for k := range originalEnv {
		os.Unsetenv(k)
	}

	params, err := parseEnv()
	if err != nil {
		t.Fatalf("parseEnv failed: %v", err)
	}
	if params.TimeoutSeconds != defaultConfig.timeoutSeconds {
		t.Errorf("Expected default timeout %d, got %d", defaultConfig.timeoutSeconds, params.TimeoutSeconds)
	}
	if params.MaxRetries != defaultConfig.maxRetries {
		t.Errorf("Expected default retries %d, got %d", defaultConfig.maxRetries, params.MaxRetries)
	}
	if params.Multiplier != defaultConfig.multiplier {
		t.Errorf("Expected default multiplier %f, got %f", defaultConfig.multiplier, params.Multiplier)
	}
	if params.WaitIntervalSeconds != defaultConfig.waitIntervalSeconds {
		t.Errorf("Expected default wait interval %d, got %d", defaultConfig.waitIntervalSeconds, params.WaitIntervalSeconds)
	}

	// Test with environment variables set
	_ = os.Setenv(EnvVarPrefix+EnvTimeoutSeconds, "60")
	_ = os.Setenv(EnvVarPrefix+EnvMaxRetries, "10")
	_ = os.Setenv(EnvVarPrefix+EnvWaitIntervalSeconds, "30")
	_ = os.Setenv(EnvVarPrefix+EnvBackoffMultiplier, "3.0")

	params, err = parseEnv()
	if err != nil {
		t.Fatalf("parseEnv failed: %v", err)
	}
	if params.TimeoutSeconds != 60 {
		t.Errorf("Expected timeout 60, got %d", params.TimeoutSeconds)
	}
	if params.MaxRetries != 10 {
		t.Errorf("Expected retries 10, got %d", params.MaxRetries)
	}
	if params.Multiplier != 3.0 {
		t.Errorf("Expected multiplier 3.0, got %f", params.Multiplier)
	}
	if params.WaitIntervalSeconds != 30 {
		t.Errorf("Expected wait interval 30, got %d", params.WaitIntervalSeconds)
	}
}

func TestDefaultConfigValues(t *testing.T) {
	// Test that default values are reasonable
	if defaultConfig.timeoutSeconds != 10 {
		t.Errorf("Expected default timeout 10, got %d", defaultConfig.timeoutSeconds)
	}
	if defaultConfig.maxRetries != 3 {
		t.Errorf("Expected default retries 3, got %d", defaultConfig.maxRetries)
	}
	if defaultConfig.multiplier != 2.0 {
		t.Errorf("Expected default multiplier 2.0, got %f", defaultConfig.multiplier)
	}
	if defaultConfig.waitIntervalSeconds != 10 {
		t.Errorf("Expected default wait interval 10, got %d", defaultConfig.waitIntervalSeconds)
	}
}

func TestConstantsValues(t *testing.T) {
	// Test constants
	if V4ClientDefaultPageSize != 100 {
		t.Errorf("Expected page size 100, got %d", V4ClientDefaultPageSize)
	}
	if EnvVarPrefix != "OWNERSHIT_" {
		t.Errorf("Expected prefix 'OWNERSHIT_', got '%s'", EnvVarPrefix)
	}
	if EnvTimeoutSeconds != "TIMEOUT_SECONDS" {
		t.Errorf("Expected 'TIMEOUT_SECONDS', got '%s'", EnvTimeoutSeconds)
	}
	if EnvMaxRetries != "MAX_RETRIES" {
		t.Errorf("Expected 'MAX_RETRIES', got '%s'", EnvMaxRetries)
	}
	if EnvWaitIntervalSeconds != "WAIT_INTERVAL_SECONDS" {
		t.Errorf("Expected 'WAIT_INTERVAL_SECONDS', got '%s'", EnvWaitIntervalSeconds)
	}
	if EnvBackoffMultiplier != "BACKOFF_MULTIPLIER" {
		t.Errorf("Expected 'BACKOFF_MULTIPLIER', got '%s'", EnvBackoffMultiplier)
	}
}

func TestParseEnv_ErrorPaths(t *testing.T) {
	// This tests the error handling paths in parseEnv
	// We can't easily test the log.Fatal paths without changing the logging setup
	// but we can test that the parsing works with edge cases

	// Save original environment
	originalEnv := map[string]string{
		EnvVarPrefix + EnvTimeoutSeconds:      os.Getenv(EnvVarPrefix + EnvTimeoutSeconds),
		EnvVarPrefix + EnvMaxRetries:          os.Getenv(EnvVarPrefix + EnvMaxRetries),
		EnvVarPrefix + EnvWaitIntervalSeconds: os.Getenv(EnvVarPrefix + EnvWaitIntervalSeconds),
		EnvVarPrefix + EnvBackoffMultiplier:   os.Getenv(EnvVarPrefix + EnvBackoffMultiplier),
	}

	// Clean up after test
	defer func() {
		for k, v := range originalEnv {
			if v == "" {
				os.Unsetenv(k)
			} else {
				_ = os.Setenv(k, v)
			}
		}
	}()

	// Test empty string values (should use defaults)
	_ = os.Setenv(EnvVarPrefix+EnvTimeoutSeconds, "")
	_ = os.Setenv(EnvVarPrefix+EnvMaxRetries, "")
	_ = os.Setenv(EnvVarPrefix+EnvWaitIntervalSeconds, "")
	_ = os.Setenv(EnvVarPrefix+EnvBackoffMultiplier, "")

	params, err := parseEnv()
	if err != nil {
		t.Fatalf("parseEnv failed: %v", err)
	}
	if params.TimeoutSeconds != defaultConfig.timeoutSeconds {
		t.Errorf("Expected default timeout with empty string, got %d", params.TimeoutSeconds)
	}
	if params.MaxRetries != defaultConfig.maxRetries {
		t.Errorf("Expected default retries with empty string, got %d", params.MaxRetries)
	}
	if params.Multiplier != defaultConfig.multiplier {
		t.Errorf("Expected default multiplier with empty string, got %f", params.Multiplier)
	}
	if params.WaitIntervalSeconds != defaultConfig.waitIntervalSeconds {
		t.Errorf("Expected default wait interval with empty string, got %d", params.WaitIntervalSeconds)
	}
}

func TestTypeAliasesUsage(t *testing.T) {
	// Test that type aliases are properly defined and can be used
	var teams OrganizationTeams
	if teams == nil {
		teams = OrganizationTeams{}
	}
	if len(teams) != 0 {
		t.Error("Expected empty teams slice")
	}

	var rateLimit RateLimit
	_ = rateLimit // Just test that it compiles

	var label Label
	_ = label // Just test that it compiles

	// Test that we can assign to these types
	teams = make(OrganizationTeams, 0)
	// teams will never be nil after make, so just verify it's initialized
	if len(teams) != 0 {
		t.Error("Expected empty teams slice")
	}
}
