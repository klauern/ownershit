#!/usr/bin/env python3
# /// script
# requires-python = ">=3.11"
# dependencies = [
#     "pygithub>=2.1.1",
#     "pyyaml>=6.0.1",
#     "requests>=2.31.0",
# ]
# ///

"""
Backfill repository feature settings based on actual usage.

This script checks if repositories actually have:
- Wiki pages (conservative: assumes enabled wikis have content)
- Issues (open or closed, excluding PRs)
- Projects (v1 or v2)

It then updates repositories.yaml with explicit settings for repos that
differ from the defaults.

Note: The script bases override decisions on whether features are ENABLED
in the repository settings, not on whether content exists. This prevents
accidentally disabling features that are enabled but currently unused.
"""

import copy
import os
import sys
from pathlib import Path
from typing import Optional

import requests
import yaml
from github import Auth, Github, GithubException


def check_repo_features(repo) -> dict:
    """Check if a repository actually uses wiki, issues, and projects."""
    features = {
        'has_wiki': False,
        'has_issues': False,
        'has_projects': False,
        'wiki_enabled': repo.has_wiki,
        'issues_enabled': repo.has_issues,
        'projects_enabled': repo.has_projects,
    }

    # Check if wiki has actual pages (not just enabled)
    # GitHub wikis are stored as separate git repositories
    # We can check if the wiki has content by making a HEAD request to the wiki repo
    if repo.has_wiki:
        try:
            # Get the token from the GitHub client's auth
            token = repo._requester.auth.token

            # The wiki is accessible as a git repository at owner/repo.wiki
            # We'll check if we can access the wiki home page via the web
            wiki_url = f"https://github.com/{repo.full_name}/wiki"

            # Make an authenticated request
            headers = {
                'Authorization': f'token {token}',
                'User-Agent': 'ownershit-backfill-script'
            }

            response = requests.get(wiki_url, headers=headers, timeout=10, allow_redirects=True)

            # If the response contains "Create the first page", wiki is empty
            # If it contains actual content, wiki has pages
            if response.status_code == 200:
                # Check if it's the empty wiki page
                # Empty wikis show "Create the first page" button
                is_empty = 'Create the first page' in response.text
                features['has_wiki'] = not is_empty
            else:
                # If we can't access it, be conservative
                features['has_wiki'] = True

        except (requests.RequestException, AttributeError, KeyError):
            # On any error, be conservative and assume wiki might have content
            # AttributeError: if _requester.auth.token doesn't exist
            # KeyError: if token extraction fails
            # RequestException: network/HTTP errors
            features['has_wiki'] = True

    # Check if there are any issues (open or closed)
    if repo.has_issues:
        try:
            # Use open_issues_count which includes both open issues and PRs
            # Then check if we can find at least one actual issue
            # This is more efficient than paginating through all issues
            if repo.open_issues_count > 0:
                # Try to find at least one non-PR issue
                # We'll check multiple pages if needed to avoid false negatives
                found_issue = False
                max_pages = 3  # Check up to 3 pages (90 items) before giving up
                for page_num in range(max_pages):
                    try:
                        issues = list(repo.get_issues(state='all').get_page(page_num))
                        if not issues:  # No more pages
                            break
                        # Filter out pull requests (they show up in issues API)
                        actual_issues = [i for i in issues if i.pull_request is None]
                        if actual_issues:
                            found_issue = True
                            break
                    except GithubException:
                        break
                features['has_issues'] = found_issue
            else:
                features['has_issues'] = False
        except GithubException:
            features['has_issues'] = False

    # Check if there are any projects
    if repo.has_projects:
        try:
            projects = list(repo.get_projects(state='all'))
            features['has_projects'] = len(projects) > 0
        except GithubException:
            features['has_projects'] = False

    return features


def load_config(config_path: str) -> dict:
    """
    Load the repositories.yaml configuration from disk.

    Args:
        config_path: Path to the YAML configuration file.

    Returns:
        Dictionary containing the parsed configuration.

    Raises:
        FileNotFoundError: If the config file doesn't exist.
        yaml.YAMLError: If the YAML is malformed.
    """
    with open(config_path, 'r') as f:
        return yaml.safe_load(f)


def save_config(config_path: str, config: dict) -> None:
    """
    Save the updated configuration to disk.

    Args:
        config_path: Path where the YAML configuration should be saved.
        config: Configuration dictionary to serialize.

    Note:
        Uses default_flow_style=False for readable block-style YAML,
        sort_keys=False to preserve key order, and width=120 for line wrapping.
    """
    with open(config_path, 'w') as f:
        yaml.dump(config, f, default_flow_style=False, sort_keys=False, width=120)


def should_override_default(feature_enabled: bool, default_value: Optional[bool]) -> bool:
    """
    Determine if we need to explicitly set a value to override the default.

    Args:
        feature_enabled: Whether the feature is enabled in the repository.
        default_value: The default value configured at the organization level,
                      or None if no default is set.

    Returns:
        True if an explicit setting should be added to override the default,
        False otherwise.

    Logic:
        - If no default is set (None), add explicit setting only if enabled.
        - If default exists, add explicit setting only if it differs from default.
    """
    if default_value is None:
        # No default set, only add if feature is in use
        return feature_enabled
    return feature_enabled != default_value


def should_remove_explicit_setting(feature_enabled: bool, default_value: Optional[bool]) -> bool:
    """
    Determine if we should remove an explicit setting because it matches the default.

    Args:
        feature_enabled: Whether the feature is enabled in the repository.
        default_value: The default value configured at the organization level,
                      or None if no default is set.

    Returns:
        True if the explicit setting should be removed (matches default),
        False if it should be kept.

    Logic:
        - If no default is set (None), keep all explicit settings.
        - If default exists and matches the enabled state, remove redundant setting.
    """
    if default_value is None:
        # No default, keep explicit settings
        return False
    return feature_enabled == default_value


def main():
    """
    Main entry point for the backfill script.

    Process:
        1. Validate GITHUB_TOKEN environment variable
        2. Load repositories.yaml configuration
        3. For each repository, check if features are enabled
        4. Add/remove explicit settings based on defaults
        5. Create backup and save updated configuration

    Exit codes:
        0: Success (with or without updates)
        1: Error (missing token, config not found, API errors, etc.)
    """
    # Check for GitHub token
    token = os.environ.get('GITHUB_TOKEN')
    if not token:
        print("Error: GITHUB_TOKEN environment variable not set", file=sys.stderr)
        sys.exit(1)

    # Get config file path
    config_path = sys.argv[1] if len(sys.argv) > 1 else 'repositories.yaml'

    if not Path(config_path).exists():
        print(f"Error: Config file {config_path} not found", file=sys.stderr)
        sys.exit(1)

    print(f"Loading configuration from {config_path}")
    config = load_config(config_path)

    # Create backup of original config before any mutations
    backup_path = f"{config_path}.backup"
    original_config = copy.deepcopy(config)

    org_name = config.get('organization')
    if not org_name:
        print("Error: No organization specified in config", file=sys.stderr)
        sys.exit(1)

    # Get defaults from config
    default_wiki = config.get('default_wiki')
    default_issues = config.get('default_issues')
    default_projects = config.get('default_projects')

    print(f"Defaults: wiki={default_wiki}, issues={default_issues}, projects={default_projects}")
    print(f"Checking repositories in {org_name}...")

    # Initialize GitHub client with modern auth
    auth = Auth.Token(token)
    g = Github(auth=auth)
    org = g.get_organization(org_name)

    repositories = config.get('repositories', [])
    updated_count = 0
    skipped_count = 0
    error_count = 0

    for i, repo_config in enumerate(repositories):
        repo_name = repo_config.get('name')
        if not repo_name:
            continue

        has_wiki_setting = 'wiki' in repo_config
        has_issues_setting = 'issues' in repo_config
        has_projects_setting = 'projects' in repo_config

        try:
            repo = org.get_repo(repo_name)
            features = check_repo_features(repo)

            # Extract enabled flags for decision logic
            wiki_enabled = features['wiki_enabled']
            issues_enabled = features['issues_enabled']
            projects_enabled = features['projects_enabled']

            updated = False
            changes = []

            # Handle wiki setting
            # Use wiki_enabled (not has_wiki) to avoid disabling enabled-but-unused features
            if has_wiki_setting:
                # If explicit setting matches default, remove it
                if should_remove_explicit_setting(wiki_enabled, default_wiki):
                    del repo_config['wiki']
                    changes.append("removed wiki (matches default)")
                    updated = True
            else:
                # Add explicit setting if it differs from default
                if should_override_default(wiki_enabled, default_wiki):
                    repo_config['wiki'] = wiki_enabled
                    changes.append(f"wiki={wiki_enabled}")
                    updated = True

            # Handle issues setting
            # Use issues_enabled (not has_issues) to avoid disabling enabled-but-unused features
            if has_issues_setting:
                # If explicit setting matches default, remove it
                if should_remove_explicit_setting(issues_enabled, default_issues):
                    del repo_config['issues']
                    changes.append("removed issues (matches default)")
                    updated = True
            else:
                # Add explicit setting if it differs from default
                if should_override_default(issues_enabled, default_issues):
                    repo_config['issues'] = issues_enabled
                    changes.append(f"issues={issues_enabled}")
                    updated = True

            # Handle projects setting
            # Use projects_enabled (not has_projects) to avoid disabling enabled-but-unused features
            if has_projects_setting:
                # If explicit setting matches default, remove it
                if should_remove_explicit_setting(projects_enabled, default_projects):
                    del repo_config['projects']
                    changes.append("removed projects (matches default)")
                    updated = True
            else:
                # Add explicit setting if it differs from default
                if should_override_default(projects_enabled, default_projects):
                    repo_config['projects'] = projects_enabled
                    changes.append(f"projects={projects_enabled}")
                    updated = True

            if updated:
                print(f"  [{i+1}/{len(repositories)}] {repo_name}: {', '.join(changes)}")
                updated_count += 1
            else:
                print(f"  [{i+1}/{len(repositories)}] {repo_name}: matches defaults, no changes needed")

        except GithubException as e:
            print(f"  [{i+1}/{len(repositories)}] {repo_name}: ERROR - {e.status} {e.data.get('message', 'Unknown error')}")
            error_count += 1
        except (AttributeError, KeyError, ValueError) as e:
            # Catch specific exceptions for config/data issues
            print(f"  [{i+1}/{len(repositories)}] {repo_name}: ERROR - Invalid data: {e}")
            error_count += 1

    # Save updated configuration
    if updated_count > 0:
        print(f"\nSaving backup of original configuration to {backup_path}")
        save_config(backup_path, original_config)

        print(f"Saving updated configuration to {config_path}")
        save_config(config_path, config)
        print(f"\n✅ Updated {updated_count} repositories")
    else:
        print("\n✅ No updates needed")

    print("\nSummary:")
    print(f"  Updated: {updated_count}")
    print(f"  Skipped: {skipped_count}")
    print(f"  Errors: {error_count}")
    print(f"  Total: {len(repositories)}")

    if error_count > 0:
        print("\n⚠️  Some repositories had errors. Check the output above.")
        sys.exit(1)


if __name__ == '__main__':
    main()
