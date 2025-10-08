#!/usr/bin/env python3
# /// script
# requires-python = ">=3.11"
# dependencies = [
#     "pygithub>=2.1.1",
#     "pyyaml>=6.0.1",
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

Note: Wiki detection is conservative - if a wiki is enabled, we assume it
has content since GitHub's API doesn't easily expose whether wikis have pages.
"""

import copy
import os
import sys
from pathlib import Path
from typing import Optional

import requests
import yaml
from github import Auth, Github, GithubException


def check_repo_features(repo, token: str) -> dict:
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

        except requests.RequestException:
            # On any HTTP error, be conservative and assume wiki might have content
            features['has_wiki'] = True

    # Check if there are any issues (open or closed)
    if repo.has_issues:
        try:
            # Check if at least one non-PR issue exists
            # Iterate through issues to handle pagination properly
            issues_iter = repo.get_issues(state='all')
            for issue in issues_iter:
                if issue.pull_request is None:
                    features['has_issues'] = True
                    break
            # If loop completes without break, no issues found (remains False from init)
        except GithubException:
            features['has_issues'] = False

    # Check if there are any projects
    # Note: This only detects Projects (classic), not Projects (beta/v2)
    if repo.has_projects:
        try:
            projects = list(repo.get_projects(state='all'))
            features['has_projects'] = len(projects) > 0
        except GithubException:
            features['has_projects'] = False

    return features


def load_config(config_path: str) -> dict:
    """Load the repositories.yaml configuration."""
    with open(config_path, 'r') as f:
        return yaml.safe_load(f)


def save_config(config_path: str, config: dict) -> None:
    """Save the updated configuration."""
    with open(config_path, 'w') as f:
        yaml.dump(config, f, default_flow_style=False, sort_keys=False, width=120)


def should_override_default(feature_in_use: bool, default_value: Optional[bool]) -> bool:
    """Determine if we need to explicitly set a value to override the default."""
    if default_value is None:
        # No default set, only add if feature is in use
        return feature_in_use
    return feature_in_use != default_value


def should_remove_explicit_setting(feature_in_use: bool, default_value: Optional[bool]) -> bool:
    """Determine if we should remove an explicit setting because it matches the default."""
    if default_value is None:
        # No default, keep explicit settings
        return False
    return feature_in_use == default_value


def main():
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
            skipped_count += 1
            continue

        has_wiki_setting = 'wiki' in repo_config
        has_issues_setting = 'issues' in repo_config
        has_projects_setting = 'projects' in repo_config

        try:
            repo = org.get_repo(repo_name)
            features = check_repo_features(repo, token)

            updated = False
            changes = []

            # Handle wiki setting
            if has_wiki_setting:
                # If explicit setting matches default, remove it
                if should_remove_explicit_setting(features['has_wiki'], default_wiki):
                    del repo_config['wiki']
                    changes.append("removed wiki (matches default)")
                    updated = True
            else:
                # Add explicit setting if it differs from default
                if should_override_default(features['has_wiki'], default_wiki):
                    repo_config['wiki'] = features['has_wiki']
                    changes.append(f"wiki={features['has_wiki']}")
                    updated = True

            # Handle issues setting
            if has_issues_setting:
                # If explicit setting matches default, remove it
                if should_remove_explicit_setting(features['has_issues'], default_issues):
                    del repo_config['issues']
                    changes.append("removed issues (matches default)")
                    updated = True
            else:
                # Add explicit setting if it differs from default
                if should_override_default(features['has_issues'], default_issues):
                    repo_config['issues'] = features['has_issues']
                    changes.append(f"issues={features['has_issues']}")
                    updated = True

            # Handle projects setting
            if has_projects_setting:
                # If explicit setting matches default, remove it
                if should_remove_explicit_setting(features['has_projects'], default_projects):
                    del repo_config['projects']
                    changes.append("removed projects (matches default)")
                    updated = True
            else:
                # Add explicit setting if it differs from default
                if should_override_default(features['has_projects'], default_projects):
                    repo_config['projects'] = features['has_projects']
                    changes.append(f"projects={features['has_projects']}")
                    updated = True

            if updated:
                print(f"  [{i+1}/{len(repositories)}] {repo_name}: {', '.join(changes)}")
                updated_count += 1
            else:
                print(f"  [{i+1}/{len(repositories)}] {repo_name}: matches defaults, no changes needed")

        except GithubException as e:
            print(f"  [{i+1}/{len(repositories)}] {repo_name}: ERROR - {e.status} {e.data.get('message', 'Unknown error')}")
            error_count += 1
        except Exception as e:
            # Catch-all for unexpected errors
            print(f"  [{i+1}/{len(repositories)}] {repo_name}: ERROR - {e}")
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
