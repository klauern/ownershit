package ownershit

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

// getCSVHeaders returns the standardized CSV column headers used when exporting repository
// configuration. The returned slice lists columns in the exact order expected by import/export
// routines (owner, repo, organization, wiki_enabled, ..., push_allowlist).
func getCSVHeaders() []string {
	return []string{
		"owner", "repo", "organization", "wiki_enabled", "issues_enabled",
		"projects_enabled", "private", "archived", "template", "default_branch",
		"delete_branch_on_merge", "discussions_enabled", "sponsorships_enabled", "require_pull_request_reviews",
		"require_approving_count", "require_code_owners", "allow_merge_commit",
		"allow_squash_merge", "allow_rebase_merge", "require_status_checks",
		"require_up_to_date_branch", "enforce_admins", "restrict_pushes",
		"require_conversation_resolution", "require_linear_history",
		"allow_force_pushes", "allow_deletions", "status_checks", "push_allowlist",
	}
}

// convertToCSVRow converts a PermissionsSettings for a single repository context into a CSV row.
// The returned slice contains columns in the same order as getCSVHeaders().
// If config is nil or contains no repository entries the function returns an
// empty row with the correct number of columns. When config.Repositories is
// populated only the first repository entry is used; branch-related values are
// taken from config.BranchPermissions. Pointer and slice fields are serialized
// using helper functions so missing values become empty strings.
func convertToCSVRow(config *PermissionsSettings, owner, repo string) []string {
	if config == nil || len(config.Repositories) == 0 {
		row := make([]string, len(getCSVHeaders()))
		row[0] = sanitizeCSV(owner) // owner
		row[1] = sanitizeCSV(repo)  // repo
		return row
	}

	repoConfig := config.Repositories[0] // Single repository context
	branchPerms := &config.BranchPermissions

	return []string{
		sanitizeCSV(owner),                                       // owner
		sanitizeCSV(repo),                                        // repo
		safeStringValue(config.Organization),                     // organization
		safeBoolValue(repoConfig.Wiki),                           // wiki_enabled
		safeBoolValue(repoConfig.Issues),                         // issues_enabled
		safeBoolValue(repoConfig.Projects),                       // projects_enabled
		safeBoolValue(repoConfig.Private),                        // private
		safeBoolValue(repoConfig.Archived),                       // archived
		safeBoolValue(repoConfig.Template),                       // template
		safeStringValue(repoConfig.DefaultBranch),                // default_branch
		safeBoolValue(repoConfig.DeleteBranchOnMerge),            // delete_branch_on_merge
		safeBoolValue(repoConfig.HasDiscussionsEnabled),          // discussions_enabled
		safeBoolValue(repoConfig.HasSponsorshipsEnabled),         // sponsorships_enabled
		safeBoolValue(branchPerms.RequirePullRequestReviews),     // require_pull_request_reviews
		safeIntValue(branchPerms.ApproverCount),                  // require_approving_count
		safeBoolValue(branchPerms.RequireCodeOwners),             // require_code_owners
		safeBoolValue(branchPerms.AllowMergeCommit),              // allow_merge_commit
		safeBoolValue(branchPerms.AllowSquashMerge),              // allow_squash_merge
		safeBoolValue(branchPerms.AllowRebaseMerge),              // allow_rebase_merge
		safeBoolValue(branchPerms.RequireStatusChecks),           // require_status_checks
		safeBoolValue(branchPerms.RequireUpToDateBranch),         // require_up_to_date_branch
		safeBoolValue(branchPerms.EnforceAdmins),                 // enforce_admins
		safeBoolValue(branchPerms.RestrictPushes),                // restrict_pushes
		safeBoolValue(branchPerms.RequireConversationResolution), // require_conversation_resolution
		safeBoolValue(branchPerms.RequireLinearHistory),          // require_linear_history
		safeBoolValue(branchPerms.AllowForcePushes),              // allow_force_pushes
		safeBoolValue(branchPerms.AllowDeletions),                // allow_deletions
		joinStringSlice(branchPerms.StatusChecks),                // status_checks
		joinStringSlice(branchPerms.PushAllowlist),               // push_allowlist
	}
}

// safeBoolValue returns an empty string if ptr is nil; otherwise it returns
// "true" or "false" corresponding to the boolean value pointed to by ptr.
func safeBoolValue(ptr *bool) string {
	if ptr == nil {
		return ""
	}
	return strconv.FormatBool(*ptr)
}

// safeStringValue returns the string pointed to by ptr, or the empty string if ptr is nil.
func safeStringValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return sanitizeCSV(*ptr)
}

// safeIntValue returns the string representation of the int pointed to by ptr.
// If ptr is nil, it returns an empty string.
func safeIntValue(ptr *int) string {
	if ptr == nil {
		return ""
	}
	return strconv.Itoa(*ptr)
}

// joinStringSlice joins the input slice with '|' and returns an empty string if the slice is empty.
func joinStringSlice(slice []string) string {
	if len(slice) == 0 {
		return ""
	}
	return sanitizeCSV(strings.Join(slice, "|"))
}

// sanitizeCSV prefixes risky values to prevent CSV formula injection in spreadsheet viewers.
func sanitizeCSV(s string) string {
	if s == "" {
		return ""
	}
	i := 0
	for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
		i++
	}
	if i < len(s) {
		switch s[i] {
		case '=', '+', '-', '@':
			return "'" + s
		}
	}
	return s
}

// ProcessRepositoriesCSV writes CSV rows for each repository in repos to the provided output writer.
// It optionally writes a header row when writeHeader is true.
// Repositories must be in the "owner/repo" format; entries that do not match are counted as failures.
// Processing continues on per-repository errors; any failures are collected and returned as a *BatchProcessingError
// that includes totals and per-repo errors. Returns nil if all repositories were processed successfully.
func ProcessRepositoriesCSV(repos []string, output io.Writer, client *GitHubClient, writeHeader bool) error {
	csvWriter := csv.NewWriter(output)

	// Write header if required
	if writeHeader {
		if err := csvWriter.Write(getCSVHeaders()); err != nil {
			return fmt.Errorf("failed to write CSV header: %w", err)
		}
	}

	var (
		successCount = 0
		errorCount   = 0
		repoErrors   []RepositoryError
	)

	log.Info().
		Int("totalRepositories", len(repos)).
		Msg("Starting CSV export for repositories")

	for i, repo := range repos {
		parts := strings.Split(repo, "/")
		if len(parts) != ownerRepoParts {
			errorCount++
			repoErr := RepositoryError{
				Repository: repo,
				Error:      fmt.Errorf("%w", ErrInvalidRepoFormat),
			}
			repoErrors = append(repoErrors, repoErr)
			continue
		}

		owner, repoName := parts[0], parts[1]

		log.Debug().
			Str("owner", owner).
			Str("repo", repoName).
			Int("progress", i+1).
			Int("total", len(repos)).
			Msg("Processing repository")

		// Process individual repository with error handling
		if err := processRepositoryToCSV(owner, repoName, csvWriter, client); err != nil {
			errorCount++
			repoErr := RepositoryError{
				Owner:      owner,
				Repository: repoName,
				Error:      err,
			}
			repoErrors = append(repoErrors, repoErr)

			// Log error but continue processing
			log.Warn().
				Err(err).
				Str("owner", owner).
				Str("repo", repoName).
				Msg("Failed to process repository")

			continue
		}

		successCount++

		// Progress logging every 10 repositories
		if (i+1)%10 == 0 {
			log.Info().
				Int("processed", i+1).
				Int("total", len(repos)).
				Int("success", successCount).
				Int("errors", errorCount).
				Msg("CSV export progress")
		}
	}

	// Final summary
	log.Info().
		Int("totalProcessed", len(repos)).
		Int("successful", successCount).
		Int("failed", errorCount).
		Msg("CSV export completed")

	// Flush and check for any buffered I/O errors
	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return fmt.Errorf("failed to flush CSV output: %w", err)
	}

	// Report errors if any occurred
	if len(repoErrors) > 0 {
		return &BatchProcessingError{
			TotalRepositories: len(repos),
			SuccessCount:      successCount,
			ErrorCount:        errorCount,
			Errors:            repoErrors,
		}
	}

	return nil
}

// processRepositoryToCSV imports configuration for the given repository, converts
// it into a CSV row, and writes it using the provided csv.Writer.
// It returns an error if importing the repository configuration fails or if writing the CSV row fails.
func processRepositoryToCSV(owner, repo string, writer *csv.Writer, client *GitHubClient) error {
	// Import repository configuration using existing function
	// For CSV export, relax team permission fetch errors to avoid aborting export on transient failures.
	config, err := ImportRepositoryConfig(owner, repo, client, true)
	if err != nil {
		return fmt.Errorf("failed to import repository configuration for %s/%s: %w", owner, repo, err)
	}

	// Convert to CSV row
	row := convertToCSVRow(config, owner, repo)

	// Write to CSV with error handling
	if err := writer.Write(row); err != nil {
		return fmt.Errorf("failed to write CSV row: %w", err)
	}

	return nil
}

// ParseRepositoryList parses a list of repositories from command-line arguments and an optional batch file.
//
// It validates each repository string to ensure it follows the `owner/repo` format, appends repositories read
// from the provided batch file (if any), and returns a deduplicated list preserving the order of first occurrence.
// If any command-line argument is invalid the function returns an error immediately. If reading the batch file fails,
// an error describing that failure is returned.
func ParseRepositoryList(args []string, batchFile string) ([]string, error) {
	var repos []string

	// Add command line arguments
	for _, arg := range args {
		arg = strings.TrimSpace(arg)
		if err := validateRepoFormat(arg); err != nil {
			return nil, fmt.Errorf("invalid repository format '%s': %w", arg, err)
		}
		repos = append(repos, arg)
	}

	// Add repositories from batch file
	if batchFile != "" {
		batchRepos, err := readBatchFile(batchFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read batch file: %w", err)
		}
		repos = append(repos, batchRepos...)
	}

	// Remove duplicates
	return removeDuplicates(repos), nil
}

// validateRepoFormat verifies that the given repository string is in the form "owner/repo".
// It returns nil for a valid string and an error if the string does not contain exactly two
// non-empty components separated by a single '/'.
const ownerRepoParts = 2

func validateRepoFormat(repo string) error {
	parts := strings.Split(repo, "/")
	if len(parts) != ownerRepoParts || parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("%w", ErrInvalidRepoFormat)
	}
	return nil
}

// readBatchFile reads repository entries from the named file and returns them as a slice of
// strings in the order found. The file is parsed line-by-line; blank lines and lines starting
// with `#` are ignored. Each non-empty line must be a valid "owner/repo" string — otherwise
// an error is returned identifying the offending line number and content. Any I/O or scanning
// error encountered while reading the file is returned.
func readBatchFile(filename string) ([]string, error) {
	file, err := os.Open(filename) // #nosec G304 - filename is validated by caller
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			log.Error().Err(closeErr).Str("filename", filename).Msg("failed to close file")
		}
	}()

	var repos []string
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if err := validateRepoFormat(line); err != nil {
			return nil, fmt.Errorf("line %d: invalid format '%s': %w", lineNum, line, err)
		}

		repos = append(repos, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return repos, nil
}

// removeDuplicates returns a new slice containing only the first occurrence of each repository
// from the input slice, preserving the original order and removing subsequent duplicates. It
// compares entries by exact string equality.
func removeDuplicates(repos []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, repo := range repos {
		if !seen[repo] {
			seen[repo] = true
			result = append(result, repo)
		}
	}

	return result
}

// ValidateCSVAppendMode validates that an existing CSV file (if present and non-empty)
// has headers that exactly match the expected CSV export headers.
// It returns nil if the file does not exist, is empty, or the headers match.
// Returns an error if the file cannot be opened, the headers cannot be read, or the
// existing headers are incompatible with the expected format.
func ValidateCSVAppendMode(filename string) error {
	file, err := os.Open(filename) // #nosec G304 - filename is validated by caller
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist, no validation needed
		}
		return err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			log.Error().Err(closeErr).Str("filename", filename).Msg("failed to close file")
		}
	}()

	reader := csv.NewReader(file)
	headers, err := reader.Read()
	if err != nil {
		if err == io.EOF {
			return nil // Empty file, no validation needed
		}
		return fmt.Errorf("failed to read existing CSV headers: %w", err)
	}

	expectedHeaders := getCSVHeaders()
	if !sliceEqual(headers, expectedHeaders) {
		return fmt.Errorf("%w.\nexpected: %q\ngot:      %q", ErrIncompatibleCSVHeaders, expectedHeaders, headers)
	}

	return nil
}

// sliceEqual reports whether two string slices are equal — they have the same length
// and identical elements in the same order.
func sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// RepositoryError represents an error processing a specific repository.
type RepositoryError struct {
	Owner      string
	Repository string
	Error      error
}

// BatchProcessingError represents errors from batch processing operations.
type BatchProcessingError struct {
	TotalRepositories int
	SuccessCount      int
	ErrorCount        int
	Errors            []RepositoryError
}

func (e *BatchProcessingError) Error() string {
	return fmt.Sprintf("batch processing completed: %d of %d failed (%d successful, %d failed)",
		e.ErrorCount, e.TotalRepositories, e.SuccessCount, e.ErrorCount)
}

// GetDetailedErrors returns detailed error messages for each failed repository.
func (e *BatchProcessingError) GetDetailedErrors() []string {
	var details []string
	for _, repoErr := range e.Errors {
		if repoErr.Owner != "" {
			details = append(details, fmt.Sprintf("%s/%s: %s",
				repoErr.Owner, repoErr.Repository, repoErr.Error.Error()))
		} else {
			details = append(details, fmt.Sprintf("%s: %s",
				repoErr.Repository, repoErr.Error.Error()))
		}
	}
	return details
}

// Package-level errors for consistent wrapping and lint compliance.
var (
	ErrInvalidRepoFormat      = errors.New("invalid repository format, must be 'owner/repo'")
	ErrIncompatibleCSVHeaders = errors.New("existing CSV has incompatible headers")
)
