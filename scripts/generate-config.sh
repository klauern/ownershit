#!/bin/bash
# Generate ownershit configuration from GitHub user repositories
# 
# Usage: ./scripts/generate-config.sh <github-username> [output-file]
#
# Example:
#   ./scripts/generate-config.sh klauern klauern-repositories.yaml
#
# Requirements:
#   - GITHUB_TOKEN environment variable must be set
#   - Go 1.21+ must be installed

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Check for GitHub token
if [ -z "$GITHUB_TOKEN" ]; then
    echo "Error: GITHUB_TOKEN environment variable not set"
    echo ""
    echo "Please set your GitHub Personal Access Token:"
    echo "  export GITHUB_TOKEN=ghp_your_token_here"
    exit 1
fi

# Check for username argument
if [ $# -lt 1 ]; then
    echo "Usage: $0 <github-username> [output-file]"
    echo ""
    echo "Example:"
    echo "  $0 klauern klauern-repositories.yaml"
    exit 1
fi

USERNAME="$1"
OUTPUT_FILE="${2:-${USERNAME}-repositories.yaml}"

echo "Generating configuration for user: $USERNAME"
echo "Output file: $OUTPUT_FILE"
echo ""

# Run the Go script
cd "$PROJECT_ROOT"
go run "$SCRIPT_DIR/generate-user-config.go" "$USERNAME" "$OUTPUT_FILE"

echo ""
echo "Next steps:"
echo "  1. Review the generated file: $OUTPUT_FILE"
echo "  2. Add team permissions in the 'team:' section"
echo "  3. Configure branch protection rules in the 'branches:' section"
echo "  4. Add default labels and topics as needed"
echo "  5. Test with dry-run: ownershit sync --config $OUTPUT_FILE --dry-run"
echo "  6. Apply changes: ownershit sync --config $OUTPUT_FILE"
