package ownershit

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

// getCSVHeaders returns the ordered CSV column headers used when exporting repository configuration.
// The headers correspond to repository metadata and branch/permission settings and must match the CSV import/export format.
func getCSVHeaders() []string {
	return []string{
		"owner", "repo", "organization", "wiki_enabled", "issues_enabled",
		"projects_enabled", "private", "archived", "template", "default_branch",
		"delete_branch_on_merge", "discussions_enabled", "require_pull_request_reviews",
		"require_approving_count", "require_code_owners", "allow_merge_commit",
		"allow_squash_merge", "allow_rebase_merge", "require_status_checks",
		"require_up_to_date_branch", "enforce_admins", "restrict_pushes",
		"require_conversation_resolution", "require_linear_history",
		"allow_force_pushes", "allow_deletions", "status_checks", "push_allowlist",
	}
}

// convertToCSVRow converts a PermissionsSettings for a single repository into a CSV row.
// If config is nil or contains no repository entries, it returns an empty row with the
// same number of columns as getCSVHeaders(). The returned slice's fields correspond to
// the columns defined by getCSVHeaders(), in the same order.
func convertToCSVRow(config *PermissionsSettings, owner, repo string) []string {
	if config == nil || len(config.Repositories) == 0 {
		// Return empty row with correct number of columns
		return make([]string, len(getCSVHeaders()))
	}

	repoConfig := config.Repositories[0] // Single repository context
	branchPerms := &config.BranchPermissions

	return []string{
		owner,                                                    // owner
		repo,                                                     // repo
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

// safeBoolValue returns an empty string if ptr is nil; otherwise it returns the boolean value as "true" or "false".
func safeBoolValue(ptr *bool) string {
	if ptr == nil {
		return ""
	}
	return strconv.FormatBool(*ptr)
}

// safeStringValue returns the string pointed to by ptr, or an empty string if ptr is nil.
func safeStringValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

// safeIntValue returns the decimal string representation of the integer pointed to by ptr.
// If ptr is nil, it returns an empty string.
func safeIntValue(ptr *int) string {
	if ptr == nil {
		return ""
	}
	return strconv.Itoa(*ptr)
}

// joinStringSlice joins the elements of slice using '|' as the delimiter.
// If slice is empty, it returns an empty string.
func joinStringSlice(slice []string) string {
	if len(slice) == 0 {
		return ""
	}
	return strings.Join(slice, "|")
}

// ProcessRepositoriesCSV processes multiple repositories and writes their configuration rows
// to the provided writer in CSV format.
//
// The repos slice must contain repository identifiers in the form "owner/repo". If writeHeader
// is true, a standardized CSV header row is written first. Each repository is processed
// independently; invalid repository formats or per-repository failures are collected and do
// not stop the overall run. If any repository fails to be processed, the function returns a
// *BatchProcessingError describing totals and per-repository errors; otherwise it returns nil.
func ProcessRepositoriesCSV(repos []string, output io.Writer, client *GitHubClient, writeHeader bool) error {
	csvWriter := csv.NewWriter(output)
	defer csvWriter.Flush()

	// Write header if required
	if writeHeader {
		if err := csvWriter.Write(getCSVHeaders()); err != nil {
			return fmt.Errorf("failed to write CSV header: %w", err)
		}
	}

	var (
		successCount = 0
		errorCount   = 0
		errors       []RepositoryError
	)

	log.Info().
		Int("totalRepositories", len(repos)).
		Msg("Starting CSV import for repositories")

	for i, repo := range repos {
		parts := strings.Split(repo, "/")
		if len(parts) != 2 {
			errorCount++
			repoErr := RepositoryError{
				Repository: repo,
				Error:      fmt.Errorf("invalid repository format, must be 'owner/repo'"),
			}
			errors = append(errors, repoErr)
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
			errors = append(errors, repoErr)

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
				Msg("Batch processing progress")
		}
	}

	// Final summary
	log.Info().
		Int("totalProcessed", len(repos)).
		Int("successful", successCount).
		Int("failed", errorCount).
		Msg("CSV import completed")

	// Report errors if any occurred
	if len(errors) > 0 {
		return &BatchProcessingError{
			TotalRepositories: len(repos),
			SuccessCount:      successCount,
			ErrorCount:        errorCount,
			Errors:            errors,
		}
	}

	return nil
}

// processRepositoryToCSV imports the repository configuration for the given owner and repo,
// converts it into a CSV row, and writes that row to the provided csv.Writer.
// Returns an error if importing the configuration or writing the CSV row fails.
func processRepositoryToCSV(owner, repo string, writer *csv.Writer, client *GitHubClient) error {
	// Import repository configuration using existing function
	config, err := ImportRepositoryConfig(owner, repo, client)
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

// ParseRepositoryList parses repository identifiers from command-line arguments and an optional
// batch file, validates each entry is in "owner/repo" format, and returns a deduplicated
// slice preserving the first occurrence order. If any command-line argument is invalid the
// function returns an error; if a batch file is provided, failures to read or parse that file
// are returned as errors.
func ParseRepositoryList(args []string, batchFile string) ([]string, error) {
	var repos []string

	// Add command line arguments
	for _, arg := range args {
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

// validateRepoFormat checks that the repo string is in the form "owner/repo"
// with non-empty owner and repository parts. It returns an error when the format
// is invalid.
func validateRepoFormat(repo string) error {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("must be in format 'owner/repo'")
	}
	return nil
}

// readBatchFile reads repository entries from the named file and returns them as a slice.
// 
// Lines that are empty or start with `#` are ignored. Each non-comment line must be a
// repository in the form `owner/repo`; lines that fail validation cause an immediate error
// that includes the offending line number and content. Any error while opening or scanning
// the file is returned wrapped.
func readBatchFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

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

// removeDuplicates returns a new slice containing the elements of repos with duplicates removed,
// preserving the original order of first occurrences.
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

// returned directly.
func ValidateCSVAppendMode(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist, no validation needed
		}
		return err
	}
	defer file.Close()

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
		return fmt.Errorf("existing CSV has incompatible headers")
	}

	return nil
}

// sliceEqual reports whether two string slices have identical length and elements in the same order.
// Nil and empty slices are treated as equal (both have length zero).
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

// RepositoryError represents an error processing a specific repository
type RepositoryError struct {
	Owner      string
	Repository string
	Error      error
}

// BatchProcessingError represents errors from batch processing operations
type BatchProcessingError struct {
	TotalRepositories int
	SuccessCount      int
	ErrorCount        int
	Errors            []RepositoryError
}

func (e *BatchProcessingError) Error() string {
	return fmt.Sprintf("batch processing completed with %d/%d failures: %d successful, %d failed",
		e.ErrorCount, e.TotalRepositories, e.SuccessCount, e.ErrorCount)
}

// GetDetailedErrors returns detailed error messages for each failed repository
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
