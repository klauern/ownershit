package ownershit

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

// Base error types for common categories.
var (
	ErrAuthentication   = errors.New("authentication error")
	ErrPermissionDenied = errors.New("permission denied")
	ErrNotFound         = errors.New("resource not found")
	ErrConfiguration    = errors.New("configuration error")
	ErrValidation       = errors.New("validation error")
	ErrNetwork          = errors.New("network error")
	ErrRateLimit        = errors.New("rate limit exceeded")
)

// GitHubAPIError represents errors from GitHub API interactions.
type GitHubAPIError struct {
	StatusCode int
	Operation  string
	Repository string
	Message    string
	Err        error
}

func (e *GitHubAPIError) Error() string {
	if e.Repository != "" {
		return fmt.Sprintf("GitHub API error [%d] for %s on %s: %s", e.StatusCode, e.Operation, e.Repository, e.Message)
	}
	return fmt.Sprintf("GitHub API error [%d] for %s: %s", e.StatusCode, e.Operation, e.Message)
}

func (e *GitHubAPIError) Unwrap() error {
	return e.Err
}

// Is checks if the error matches based on status code and base error types.
func (e *GitHubAPIError) Is(target error) bool {
	switch {
	case e.StatusCode == http.StatusUnauthorized && errors.Is(target, ErrAuthentication):
		return true
	case e.StatusCode == http.StatusForbidden && errors.Is(target, ErrPermissionDenied):
		return true
	case e.StatusCode == http.StatusNotFound && errors.Is(target, ErrNotFound):
		return true
	case e.StatusCode == http.StatusTooManyRequests && errors.Is(target, ErrRateLimit):
		return true
	}
	return false
}

// NewGitHubAPIError creates a new GitHub API error.
func NewGitHubAPIError(statusCode int, operation, repository, message string, err error) *GitHubAPIError {
	return &GitHubAPIError{
		StatusCode: statusCode,
		Operation:  operation,
		Repository: repository,
		Message:    message,
		Err:        err,
	}
}

// AuthenticationError represents authentication-related errors.
type AuthenticationError struct {
	TokenType string
	Message   string
	Err       error
}

func (e *AuthenticationError) Error() string {
	if e.TokenType != "" {
		return fmt.Sprintf("authentication failed for %s token: %s", e.TokenType, e.Message)
	}
	return fmt.Sprintf("authentication failed: %s", e.Message)
}

func (e *AuthenticationError) Unwrap() error {
	return e.Err
}

// Is reports whether the target error is an authentication error.
func (e *AuthenticationError) Is(target error) bool {
	return errors.Is(target, ErrAuthentication)
}

// NewAuthenticationError creates a new authentication error.
func NewAuthenticationError(tokenType, message string, err error) *AuthenticationError {
	return &AuthenticationError{
		TokenType: tokenType,
		Message:   message,
		Err:       err,
	}
}

// RateLimitError represents rate limiting errors.
type RateLimitError struct {
	ResetTime time.Time
	Remaining int
	Message   string
	Err       error
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded, %d requests remaining, resets at %s: %s",
		e.Remaining, e.ResetTime.Format(time.RFC3339), e.Message)
}

func (e *RateLimitError) Unwrap() error {
	return e.Err
}

// Is reports whether the target error is a rate limit error.
func (e *RateLimitError) Is(target error) bool {
	return errors.Is(target, ErrRateLimit)
}

// NewRateLimitError creates a new rate limit error.
func NewRateLimitError(resetTime time.Time, remaining int, message string, err error) *RateLimitError {
	return &RateLimitError{
		ResetTime: resetTime,
		Remaining: remaining,
		Message:   message,
		Err:       err,
	}
}

// RepositoryNotFoundError represents repository not found errors.
type RepositoryNotFoundError struct {
	Owner string
	Name  string
	Err   error
}

func (e *RepositoryNotFoundError) Error() string {
	return fmt.Sprintf("repository %s/%s not found", e.Owner, e.Name)
}

func (e *RepositoryNotFoundError) Unwrap() error {
	return e.Err
}

// Is reports whether the target error is a repository not found error.
func (e *RepositoryNotFoundError) Is(target error) bool {
	return errors.Is(target, ErrNotFound)
}

// NewRepositoryNotFoundError creates a new repository not found error.
func NewRepositoryNotFoundError(owner, name string, err error) *RepositoryNotFoundError {
	return &RepositoryNotFoundError{
		Owner: owner,
		Name:  name,
		Err:   err,
	}
}

// PermissionDeniedError represents permission-related errors.
type PermissionDeniedError struct {
	Operation  string
	Repository string
	Required   string
	Message    string
	Err        error
}

func (e *PermissionDeniedError) Error() string {
	if e.Required != "" {
		return fmt.Sprintf("permission denied for %s on %s (requires %s): %s",
			e.Operation, e.Repository, e.Required, e.Message)
	}
	return fmt.Sprintf("permission denied for %s on %s: %s", e.Operation, e.Repository, e.Message)
}

func (e *PermissionDeniedError) Unwrap() error {
	return e.Err
}

// Is reports whether the target error is a permission denied error.
func (e *PermissionDeniedError) Is(target error) bool {
	return errors.Is(target, ErrPermissionDenied)
}

// NewPermissionDeniedError creates a new permission denied error.
func NewPermissionDeniedError(operation, repository, required, message string, err error) *PermissionDeniedError {
	return &PermissionDeniedError{
		Operation:  operation,
		Repository: repository,
		Required:   required,
		Message:    message,
		Err:        err,
	}
}

// BranchProtectionRuleExistsError represents the case where a branch protection
// rule already exists for a given pattern. This is an expected condition when
// attempting to create a rule that’s already present and should trigger a
// fallback or update path rather than be treated as a hard error.
type BranchProtectionRuleExistsError struct {
	Pattern string
}

func (e *BranchProtectionRuleExistsError) Error() string {
	return fmt.Sprintf("branch protection rule already exists for pattern %s", e.Pattern)
}

// ConfigValidationError represents configuration validation errors.
type ConfigValidationError struct {
	Field   string
	Value   interface{}
	Message string
	Err     error
}

func (e *ConfigValidationError) Error() string {
	return fmt.Sprintf("configuration validation failed for field '%s' with value '%v': %s",
		e.Field, e.Value, e.Message)
}

func (e *ConfigValidationError) Unwrap() error {
	return e.Err
}

// Is reports whether the target error is a config validation error.
func (e *ConfigValidationError) Is(target error) bool {
	return errors.Is(target, ErrConfiguration) || errors.Is(target, ErrValidation)
}

// NewConfigValidationError creates a new configuration validation error.
func NewConfigValidationError(field string, value interface{}, message string, err error) *ConfigValidationError {
	return &ConfigValidationError{
		Field:   field,
		Value:   value,
		Message: message,
		Err:     err,
	}
}

// ConfigFileError represents configuration file-related errors.
type ConfigFileError struct {
	Filename  string
	Operation string
	Message   string
	Err       error
}

func (e *ConfigFileError) Error() string {
	return fmt.Sprintf("configuration file error during %s of %s: %s", e.Operation, e.Filename, e.Message)
}

func (e *ConfigFileError) Unwrap() error {
	return e.Err
}

// Is reports whether the target error is a config file error.
func (e *ConfigFileError) Is(target error) bool {
	return errors.Is(target, ErrConfiguration)
}

// NewConfigFileError creates a new configuration file error.
func NewConfigFileError(filename, operation, message string, err error) *ConfigFileError {
	return &ConfigFileError{
		Filename:  filename,
		Operation: operation,
		Message:   message,
		Err:       err,
	}
}

// ArchiveEligibilityError represents repository archiving eligibility errors.
type ArchiveEligibilityError struct {
	Repository string
	Reason     string
	Criteria   map[string]interface{}
	Err        error
}

func (e *ArchiveEligibilityError) Error() string {
	return fmt.Sprintf("repository %s is not eligible for archiving: %s", e.Repository, e.Reason)
}

func (e *ArchiveEligibilityError) Unwrap() error {
	return e.Err
}

// Is reports whether the target error is an archive eligibility error.
func (e *ArchiveEligibilityError) Is(target error) bool {
	return errors.Is(target, ErrValidation)
}

// NewArchiveEligibilityError creates a new archive eligibility error.
func NewArchiveEligibilityError(repository, reason string, criteria map[string]interface{}, err error) *ArchiveEligibilityError {
	return &ArchiveEligibilityError{
		Repository: repository,
		Reason:     reason,
		Criteria:   criteria,
		Err:        err,
	}
}

// NetworkError represents network-related errors.
type NetworkError struct {
	Operation string
	URL       string
	Message   string
	Err       error
}

func (e *NetworkError) Error() string {
	if e.URL != "" {
		return fmt.Sprintf("network error during %s to %s: %s", e.Operation, e.URL, e.Message)
	}
	return fmt.Sprintf("network error during %s: %s", e.Operation, e.Message)
}

func (e *NetworkError) Unwrap() error {
	return e.Err
}

// Is reports whether the target error is a network error.
func (e *NetworkError) Is(target error) bool {
	return errors.Is(target, ErrNetwork)
}

// NewNetworkError creates a new network error.
func NewNetworkError(operation, url, message string, err error) *NetworkError {
	return &NetworkError{
		Operation: operation,
		URL:       url,
		Message:   message,
		Err:       err,
	}
}
