#!/usr/bin/env -S uv run
# /// script
# requires-python = ">=3.11"
# dependencies = [
#     "pygithub>=2.1.1",
#     "pyyaml>=6.0.1",
#     "requests>=2.32.3",
# ]
# ///

"""
Backfill repository feature settings based on actual usage.

This script checks if repositories actually have:
- Wiki pages (conservative: assumes enabled wikis have content)
- Issues (open or closed, excluding PRs)
- Projects (classic/v1 only; does not detect Projects v2/beta)
- Delete branch on merge setting
- Discussions enabled
- Private/Public visibility
- Archived status
- Template repository status
- Default branch (if not 'main')
- Description and homepage

It then updates repositories.yaml with explicit settings for repos that
differ from the defaults.

Note: Wiki detection is conservative - if a wiki is enabled, we assume it
has content since GitHub's API doesn't easily expose whether wikis have pages.
"""

import os
import shutil
import sys
from pathlib import Path
from typing import Optional

import requests
import yaml
from github import Auth, Github, GithubException


def check_repo_features(repo) -> dict:
    """Check if a repository actually uses wiki, issues, projects, and other repository-level settings."""
    features = {
        'has_wiki': False,
        'has_issues': False,
        'has_projects': False,
        'delete_branch_on_merge': repo.delete_branch_on_merge if hasattr(repo, 'delete_branch_on_merge') else None,
        'wiki_enabled': repo.has_wiki,
        'issues_enabled': repo.has_issues,
        'projects_enabled': repo.has_projects,
        'discussions_enabled': repo.has_discussions if hasattr(repo, 'has_discussions') else None,
        'private': repo.private,
        'archived': repo.archived,
        'template': repo.is_template if hasattr(repo, 'is_template') else None,
        'default_branch': repo.default_branch,
        'description': repo.description if repo.description else None,
        'homepage': repo.homepage if repo.homepage else None,
    }

    # Check if wiki has actual pages (not just enabled)
    # GitHub wikis are stored as separate git repositories
    # We can check if the wiki has content by making a HEAD request to the wiki repo
    if repo.has_wiki:
        try:
            # The wiki is accessible as a git repository at owner/repo.wiki
            # We'll check if we can access the wiki home page via the web
            wiki_url = f"https://github.com/{repo.full_name}/wiki"

            # Make an unauthenticated request (PATs don't authenticate HTML endpoints)
            headers = {
                'User-Agent': 'ownershit-backfill-script',
                'Accept': 'text/html',
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

    org_name = config.get('organization')
    if not org_name:
        print("Error: No organization specified in config", file=sys.stderr)
        sys.exit(1)

    # Get defaults from config (supports both old and new format)
    defaults = config.get('defaults', {})
    default_wiki = defaults.get('wiki')
    default_issues = defaults.get('issues')
    default_projects = defaults.get('projects')
    default_delete_branch = defaults.get('delete_branch_on_merge')
    default_discussions = defaults.get('discussions_enabled')

    # Fallback to old format for backward compatibility
    if not defaults:
        default_wiki = config.get('default_wiki')
        default_issues = config.get('default_issues')
        default_projects = config.get('default_projects')
        print("Note: Using legacy default_* fields. Consider migrating to nested 'defaults' block.")

    print(f"Defaults: wiki={default_wiki}, issues={default_issues}, projects={default_projects}, delete_branch_on_merge={default_delete_branch}, discussions_enabled={default_discussions}")
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
            features = check_repo_features(repo)

            updated = False
            changes = []

            # Handle wiki setting
            if has_wiki_setting:
                current_value = repo_config.get('wiki')
                # If explicit setting differs from actual usage, update it
                if current_value != features['has_wiki']:
                    repo_config['wiki'] = features['has_wiki']
                    changes.append(f"wiki={features['has_wiki']} (was {current_value})")
                    updated = True
                # If explicit setting matches default, remove it
                elif should_remove_explicit_setting(features['has_wiki'], default_wiki):
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
                current_value = repo_config.get('issues')
                # If explicit setting differs from actual usage, update it
                if current_value != features['has_issues']:
                    repo_config['issues'] = features['has_issues']
                    changes.append(f"issues={features['has_issues']} (was {current_value})")
                    updated = True
                # If explicit setting matches default, remove it
                elif should_remove_explicit_setting(features['has_issues'], default_issues):
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
                current_value = repo_config.get('projects')
                # If explicit setting differs from actual usage, update it
                if current_value != features['has_projects']:
                    repo_config['projects'] = features['has_projects']
                    changes.append(f"projects={features['has_projects']} (was {current_value})")
                    updated = True
                # If explicit setting matches default, remove it
                elif should_remove_explicit_setting(features['has_projects'], default_projects):
                    del repo_config['projects']
                    changes.append("removed projects (matches default)")
                    updated = True
            else:
                # Add explicit setting if it differs from default
                if should_override_default(features['has_projects'], default_projects):
                    repo_config['projects'] = features['has_projects']
                    changes.append(f"projects={features['has_projects']}")
                    updated = True

            # Handle delete_branch_on_merge setting
            has_delete_branch_setting = 'delete_branch_on_merge' in repo_config
            if features['delete_branch_on_merge'] is not None:
                if has_delete_branch_setting:
                    current_value = repo_config.get('delete_branch_on_merge')
                    # If explicit setting differs from actual value, update it
                    if current_value != features['delete_branch_on_merge']:
                        repo_config['delete_branch_on_merge'] = features['delete_branch_on_merge']
                        changes.append(f"delete_branch_on_merge={features['delete_branch_on_merge']} (was {current_value})")
                        updated = True
                    # If explicit setting matches default, remove it
                    elif should_remove_explicit_setting(features['delete_branch_on_merge'], default_delete_branch):
                        del repo_config['delete_branch_on_merge']
                        changes.append("removed delete_branch_on_merge (matches default)")
                        updated = True
                else:
                    # Add explicit setting if it differs from default
                    if should_override_default(features['delete_branch_on_merge'], default_delete_branch):
                        repo_config['delete_branch_on_merge'] = features['delete_branch_on_merge']
                        changes.append(f"delete_branch_on_merge={features['delete_branch_on_merge']}")
                        updated = True

            # Handle discussions_enabled setting
            has_discussions_setting = 'discussions_enabled' in repo_config
            if features['discussions_enabled'] is not None:
                if has_discussions_setting:
                    current_value = repo_config.get('discussions_enabled')
                    if current_value != features['discussions_enabled']:
                        repo_config['discussions_enabled'] = features['discussions_enabled']
                        changes.append(f"discussions_enabled={features['discussions_enabled']} (was {current_value})")
                        updated = True
                    elif should_remove_explicit_setting(features['discussions_enabled'], default_discussions):
                        del repo_config['discussions_enabled']
                        changes.append("removed discussions_enabled (matches default)")
                        updated = True
                else:
                    if should_override_default(features['discussions_enabled'], default_discussions):
                        repo_config['discussions_enabled'] = features['discussions_enabled']
                        changes.append(f"discussions_enabled={features['discussions_enabled']}")
                        updated = True

            # Handle private setting (always set explicitly since it's important)
            has_private_setting = 'private' in repo_config
            if features['private'] is not None:
                if has_private_setting:
                    current_value = repo_config.get('private')
                    if current_value != features['private']:
                        repo_config['private'] = features['private']
                        changes.append(f"private={features['private']} (was {current_value})")
                        updated = True
                else:
                    # Always add private setting if repository is private
                    if features['private']:
                        repo_config['private'] = features['private']
                        changes.append(f"private={features['private']}")
                        updated = True

            # Handle archived setting (always set explicitly if true)
            has_archived_setting = 'archived' in repo_config
            if features['archived'] is not None:
                if has_archived_setting:
                    current_value = repo_config.get('archived')
                    if current_value != features['archived']:
                        repo_config['archived'] = features['archived']
                        changes.append(f"archived={features['archived']} (was {current_value})")
                        updated = True
                else:
                    # Always add archived setting if repository is archived
                    if features['archived']:
                        repo_config['archived'] = features['archived']
                        changes.append(f"archived={features['archived']}")
                        updated = True

            # Handle template setting (always set explicitly if true)
            has_template_setting = 'template' in repo_config
            if features['template'] is not None:
                if has_template_setting:
                    current_value = repo_config.get('template')
                    if current_value != features['template']:
                        repo_config['template'] = features['template']
                        changes.append(f"template={features['template']} (was {current_value})")
                        updated = True
                else:
                    # Always add template setting if repository is a template
                    if features['template']:
                        repo_config['template'] = features['template']
                        changes.append(f"template={features['template']}")
                        updated = True

            # Handle default_branch (always set explicitly)
            has_default_branch_setting = 'default_branch' in repo_config
            if features['default_branch'] is not None:
                # Only update if it's not 'main' (the GitHub default)
                if features['default_branch'] != 'main':
                    if has_default_branch_setting:
                        current_value = repo_config.get('default_branch')
                        if current_value != features['default_branch']:
                            repo_config['default_branch'] = features['default_branch']
                            changes.append(f"default_branch={features['default_branch']} (was {current_value})")
                            updated = True
                    else:
                        repo_config['default_branch'] = features['default_branch']
                        changes.append(f"default_branch={features['default_branch']}")
                        updated = True
                else:
                    # Remove default_branch if it's set to 'main' (the default)
                    if has_default_branch_setting:
                        del repo_config['default_branch']
                        changes.append("removed default_branch (is default 'main')")
                        updated = True

            # Handle description
            has_description_setting = 'description' in repo_config
            if features['description'] is not None:
                if has_description_setting:
                    current_value = repo_config.get('description')
                    if current_value != features['description']:
                        repo_config['description'] = features['description']
                        changes.append("description updated")
                        updated = True
                else:
                    repo_config['description'] = features['description']
                    changes.append("description added")
                    updated = True
            else:
                # Remove description if it's empty
                if has_description_setting:
                    del repo_config['description']
                    changes.append("removed empty description")
                    updated = True

            # Handle homepage
            has_homepage_setting = 'homepage' in repo_config
            if features['homepage'] is not None:
                if has_homepage_setting:
                    current_value = repo_config.get('homepage')
                    if current_value != features['homepage']:
                        repo_config['homepage'] = features['homepage']
                        changes.append("homepage updated")
                        updated = True
                else:
                    repo_config['homepage'] = features['homepage']
                    changes.append("homepage added")
                    updated = True
            else:
                # Remove homepage if it's empty
                if has_homepage_setting:
                    del repo_config['homepage']
                    changes.append("removed empty homepage")
                    updated = True

            if updated:
                print(f"  [{i+1}/{len(repositories)}] {repo_name}: {', '.join(changes)}")
                updated_count += 1
            else:
                print(f"  [{i+1}/{len(repositories)}] {repo_name}: matches defaults, no changes needed")

        except GithubException as e:
            detail = str(e)
            if isinstance(e.data, dict):
                detail = e.data.get('message', detail)
            print(f"  [{i+1}/{len(repositories)}] {repo_name}: ERROR - {e.status} {detail}")
            error_count += 1
        except Exception as e:  # noqa: BLE001
            # Catch-all for unexpected errors
            print(f"  [{i+1}/{len(repositories)}] {repo_name}: ERROR - {e}")
            error_count += 1

    # Save updated configuration
    if updated_count > 0:
        backup_path = f"{config_path}.backup"
        print(f"\nSaving backup of original configuration to {backup_path}")
        shutil.copyfile(config_path, backup_path)

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
