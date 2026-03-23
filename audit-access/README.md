# GitHub Organization Access Audit Tool

A CLI tool to audit GitHub organization membership and repository access permissions. This tool helps with governance and security auditing by providing visibility into who has elevated access across all repositories in an organization.

## Purpose

This tool is designed to help the cert-manager project (and other organizations) maintain proper governance by regularly auditing access levels. It lists all users who have triage, write, maintain, or admin permissions across repositories, excluding read-only access.

## Prerequisites

- Go 1.21 or later
- GitHub Personal Access Token with appropriate scopes

## GitHub Token Setup

You need a GitHub Personal Access Token with the following scopes:

- `repo` - Full control of private repositories (required to read collaborators)
- `read:org` - Read org and team membership, read org projects

The token can be created in: https://github.com/settings/personal-access-tokens

It must be "owned" by the org you're auditing, and you must set the token as an environment variable:

```bash
export GITHUB_TOKEN=ghp_your_token_here
```

## Usage

### Basic Usage

Audit all repositories in an organization:

```bash
export GITHUB_TOKEN=ghp_your_token_here
./audit-access org cert-manager
```

### Command-line Options

```bash
./audit-access org [organization-name] [flags]

Flags:
  -f, --format string        Output format: table, json, or csv (default "table")
  -p, --permission string    Filter by permission level: triage, write, maintain, or admin
  -r, --repo string          Audit specific repository instead of all
  -a, --include-archived     Include archived repositories
  -s, --sort-by string       Sort output by 'user' or 'repo' (default "user")
  -h, --help                 Help for org
      --version              Show version information
```

### Examples

**Audit with table output (default):**
```bash
./audit-access org cert-manager
```

**Export to JSON:**
```bash
./audit-access org cert-manager --format json > access-report.json
```

**Export to CSV:**
```bash
./audit-access org cert-manager --format csv > access-report.csv
```

**Filter by permission level:**
```bash
# Show only users with admin access
./audit-access org cert-manager --permission admin

# Show only users with write access
./audit-access org cert-manager --permission write
```

**Audit a specific repository:**
```bash
./audit-access org cert-manager --repo trust-manager
```

**Include archived repositories:**
```bash
./audit-access org cert-manager --include-archived
```

**Sort by repository instead of user:**
```bash
# Group results by repository (default is by user)
./audit-access org cert-manager --sort-by repo
```

## Output Formats

### Table (default)

```
+----------+------------------+------------+
| USERNAME | REPOSITORY       | PERMISSION |
+----------+------------------+------------+
| alice    | cert-manager     | admin      |
| alice    | trust-manager    | write      |
| bob      | cert-manager     | maintain   |
| charlie  | approver-policy  | triage     |
+----------+------------------+------------+
```

### JSON

**Sorted by user (default):**
```json
[
  {
    "username": "alice",
    "repositories": [
      {"name": "cert-manager", "permission": "admin"},
      {"name": "trust-manager", "permission": "write"}
    ]
  },
  {
    "username": "bob",
    "repositories": [
      {"name": "cert-manager", "permission": "maintain"}
    ]
  }
]
```

**Sorted by repository (`--sort-by repo`):**
```json
[
  {
    "repository": "cert-manager",
    "users": [
      {"username": "alice", "permission": "admin"},
      {"username": "bob", "permission": "maintain"}
    ]
  },
  {
    "repository": "trust-manager",
    "users": [
      {"username": "alice", "permission": "write"}
    ]
  }
]
```

### CSV

```csv
Username,Repository,Permission
alice,cert-manager,admin
alice,trust-manager,write
bob,cert-manager,maintain
charlie,approver-policy,triage
```

## Troubleshooting

### Rate Limiting

GitHub API has rate limits (5,000 requests/hour for authenticated requests). The tool displays your current rate limit status when it starts. If you hit the rate limit, wait until the reset time shown in the error message.

### Permission Denied

If you see a 403 or 404 error:

1. Verify your token has the required scopes (`repo` and `read:org`)
2. Verify you have sufficient access to the organization
