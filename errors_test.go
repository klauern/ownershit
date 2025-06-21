package ownershit

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestGitHubAPIError(t *testing.T) {
	baseErr := errors.New("base error")
	apiErr := NewGitHubAPIError(http.StatusNotFound, "get repository", "owner/repo", "not found", baseErr)

	t.Run("Error message formatting", func(t *testing.T) {
		expected := "GitHub API error [404] for get repository on owner/repo: not found"
		if apiErr.Error() != expected {
			t.Errorf("Error() = %q, want %q", apiErr.Error(), expected)
		}
	})

	t.Run("Error unwrapping", func(t *testing.T) {
		if !errors.Is(apiErr, baseErr) {
			t.Error("Expected error to unwrap to base error")
		}
	})

	t.Run("Error type matching", func(t *testing.T) {
		tests := []struct {
			statusCode int
			target     error
			want       bool
		}{
			{http.StatusUnauthorized, ErrAuthentication, true},
			{http.StatusForbidden, ErrPermissionDenied, true},
			{http.StatusNotFound, ErrNotFound, true},
			{http.StatusTooManyRequests, ErrRateLimit, true},
			{http.StatusInternalServerError, ErrNotFound, false},
		}

		for _, tt := range tests {
			err := NewGitHubAPIError(tt.statusCode, "test", "repo", "message", nil)
			if errors.Is(err, tt.target) != tt.want {
				t.Errorf("errors.Is(%d, %v) = %v, want %v", tt.statusCode, tt.target, !tt.want, tt.want)
			}
		}
	})
}

func TestAuthenticationError(t *testing.T) {
	baseErr := errors.New("token invalid")
	authErr := NewAuthenticationError("classic", "invalid token format", baseErr)

	t.Run("Error message formatting", func(t *testing.T) {
		expected := "authentication failed for classic token: invalid token format"
		if authErr.Error() != expected {
			t.Errorf("Error() = %q, want %q", authErr.Error(), expected)
		}
	})

	t.Run("Error type matching", func(t *testing.T) {
		if !errors.Is(authErr, ErrAuthentication) {
			t.Error("Expected error to match ErrAuthentication")
		}
	})

	t.Run("Error unwrapping", func(t *testing.T) {
		if !errors.Is(authErr, baseErr) {
			t.Error("Expected error to unwrap to base error")
		}
	})
}

func TestRateLimitError(t *testing.T) {
	resetTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	baseErr := errors.New("rate limited")
	rateLimitErr := NewRateLimitError(resetTime, 10, "too many requests", baseErr)

	t.Run("Error message formatting", func(t *testing.T) {
		expected := "rate limit exceeded, 10 requests remaining, resets at 2025-01-01T12:00:00Z: too many requests"
		if rateLimitErr.Error() != expected {
			t.Errorf("Error() = %q, want %q", rateLimitErr.Error(), expected)
		}
	})

	t.Run("Error type matching", func(t *testing.T) {
		if !errors.Is(rateLimitErr, ErrRateLimit) {
			t.Error("Expected error to match ErrRateLimit")
		}
	})
}

func TestRepositoryNotFoundError(t *testing.T) {
	baseErr := errors.New("404 not found")
	repoErr := NewRepositoryNotFoundError("owner", "repo", baseErr)

	t.Run("Error message formatting", func(t *testing.T) {
		expected := "repository owner/repo not found"
		if repoErr.Error() != expected {
			t.Errorf("Error() = %q, want %q", repoErr.Error(), expected)
		}
	})

	t.Run("Error type matching", func(t *testing.T) {
		if !errors.Is(repoErr, ErrNotFound) {
			t.Error("Expected error to match ErrNotFound")
		}
	})
}

func TestPermissionDeniedError(t *testing.T) {
	baseErr := errors.New("forbidden")
	permErr := NewPermissionDeniedError("push", "owner/repo", "write", "insufficient permissions", baseErr)

	t.Run("Error message formatting with required permission", func(t *testing.T) {
		expected := "permission denied for push on owner/repo (requires write): insufficient permissions"
		if permErr.Error() != expected {
			t.Errorf("Error() = %q, want %q", permErr.Error(), expected)
		}
	})

	t.Run("Error message formatting without required permission", func(t *testing.T) {
		err := NewPermissionDeniedError("push", "owner/repo", "", "insufficient permissions", nil)
		expected := "permission denied for push on owner/repo: insufficient permissions"
		if err.Error() != expected {
			t.Errorf("Error() = %q, want %q", err.Error(), expected)
		}
	})

	t.Run("Error type matching", func(t *testing.T) {
		if !errors.Is(permErr, ErrPermissionDenied) {
			t.Error("Expected error to match ErrPermissionDenied")
		}
	})
}

func TestConfigValidationError(t *testing.T) {
	baseErr := errors.New("invalid value")
	configErr := NewConfigValidationError("organization", "invalid-org", "must not be empty", baseErr)

	t.Run("Error message formatting", func(t *testing.T) {
		expected := "configuration validation failed for field 'organization' with value 'invalid-org': must not be empty"
		if configErr.Error() != expected {
			t.Errorf("Error() = %q, want %q", configErr.Error(), expected)
		}
	})

	t.Run("Error type matching", func(t *testing.T) {
		if !errors.Is(configErr, ErrConfiguration) {
			t.Error("Expected error to match ErrConfiguration")
		}
		if !errors.Is(configErr, ErrValidation) {
			t.Error("Expected error to match ErrValidation")
		}
	})
}

func TestConfigFileError(t *testing.T) {
	baseErr := errors.New("file not found")
	fileErr := NewConfigFileError("config.yaml", "read", "file does not exist", baseErr)

	t.Run("Error message formatting", func(t *testing.T) {
		expected := "configuration file error during read of config.yaml: file does not exist"
		if fileErr.Error() != expected {
			t.Errorf("Error() = %q, want %q", fileErr.Error(), expected)
		}
	})

	t.Run("Error type matching", func(t *testing.T) {
		if !errors.Is(fileErr, ErrConfiguration) {
			t.Error("Expected error to match ErrConfiguration")
		}
	})
}

func TestArchiveEligibilityError(t *testing.T) {
	criteria := map[string]interface{}{
		"stars":    0,
		"watchers": 0,
		"age":      "6 months",
	}
	baseErr := errors.New("eligibility check failed")
	archiveErr := NewArchiveEligibilityError("owner/repo", "has recent activity", criteria, baseErr)

	t.Run("Error message formatting", func(t *testing.T) {
		expected := "repository owner/repo is not eligible for archiving: has recent activity"
		if archiveErr.Error() != expected {
			t.Errorf("Error() = %q, want %q", archiveErr.Error(), expected)
		}
	})

	t.Run("Error type matching", func(t *testing.T) {
		if !errors.Is(archiveErr, ErrValidation) {
			t.Error("Expected error to match ErrValidation")
		}
	})

	t.Run("Criteria accessibility", func(t *testing.T) {
		if archiveErr.Criteria["stars"] != 0 {
			t.Error("Expected criteria to be accessible")
		}
	})
}

func TestNetworkError(t *testing.T) {
	baseErr := errors.New("connection timeout")
	networkErr := NewNetworkError("HTTP GET", "https://api.github.com", "timeout after 30s", baseErr)

	t.Run("Error message formatting with URL", func(t *testing.T) {
		expected := "network error during HTTP GET to https://api.github.com: timeout after 30s"
		if networkErr.Error() != expected {
			t.Errorf("Error() = %q, want %q", networkErr.Error(), expected)
		}
	})

	t.Run("Error message formatting without URL", func(t *testing.T) {
		err := NewNetworkError("HTTP GET", "", "timeout after 30s", nil)
		expected := "network error during HTTP GET: timeout after 30s"
		if err.Error() != expected {
			t.Errorf("Error() = %q, want %q", err.Error(), expected)
		}
	})

	t.Run("Error type matching", func(t *testing.T) {
		if !errors.Is(networkErr, ErrNetwork) {
			t.Error("Expected error to match ErrNetwork")
		}
	})
}

func TestErrorUnwrapping(t *testing.T) {
	baseErr := errors.New("original error")
	
	testErrors := []error{
		NewGitHubAPIError(404, "test", "repo", "message", baseErr),
		NewAuthenticationError("classic", "message", baseErr),
		NewRateLimitError(time.Now(), 0, "message", baseErr),
		NewRepositoryNotFoundError("owner", "repo", baseErr),
		NewPermissionDeniedError("op", "repo", "perm", "message", baseErr),
		NewConfigValidationError("field", "value", "message", baseErr),
		NewConfigFileError("file", "op", "message", baseErr),
		NewArchiveEligibilityError("repo", "reason", nil, baseErr),
		NewNetworkError("op", "url", "message", baseErr),
	}

	for i, err := range testErrors {
		t.Run(fmt.Sprintf("Error type %d unwrapping", i), func(t *testing.T) {
			if !errors.Is(err, baseErr) {
				t.Errorf("Error type %d should unwrap to base error", i)
			}
		})
	}
}