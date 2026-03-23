package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v84/github"
)

// RepoPermission represents a user's permission on a specific repository
type RepoPermission struct {
	RepoName   string `json:"name"`
	Permission string `json:"permission"`
}

// UserAccess represents a user and their repository access
type UserAccess struct {
	Username string           `json:"username"`
	Repos    []RepoPermission `json:"repositories"`
}

// FetchOrgRepositories fetches all repositories in an organization
func (c *Client) FetchOrgRepositories(ctx context.Context, org string, includeArchived bool) ([]*github.Repository, error) {
	c.logger.Info("fetching repositories", "organization", org)

	var allRepos []*github.Repository
	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		repos, resp, err := c.client.Repositories.ListByOrg(ctx, org, opt)
		if err != nil {
			return nil, c.handleRateLimit(resp, fmt.Errorf("failed to fetch repositories: %w", err))
		}

		for _, repo := range repos {
			// Skip archived repos unless explicitly requested
			if !includeArchived && repo.GetArchived() {
				continue
			}
			allRepos = append(allRepos, repo)
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	c.logger.Info("found repositories", "count", len(allRepos))
	return allRepos, nil
}

// FetchRepoCollaborators fetches all collaborators for a specific repository
func (c *Client) FetchRepoCollaborators(ctx context.Context, org, repo string) ([]*github.User, map[string]string, error) {
	var allCollaborators []*github.User
	permissions := make(map[string]string)

	opt := &github.ListCollaboratorsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		collaborators, resp, err := c.client.Repositories.ListCollaborators(ctx, org, repo, opt)
		if err != nil {
			return nil, nil, c.handleRateLimit(resp, fmt.Errorf("failed to fetch collaborators for %s: %w", repo, err))
		}

		for _, collab := range collaborators {
			allCollaborators = append(allCollaborators, collab)
			// Store the permission level for this user
			if collab.Permissions != nil {
				perm := getPermissionLevel(collab.Permissions)
				permissions[collab.GetLogin()] = perm
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allCollaborators, permissions, nil
}

// getPermissionLevel determines the highest permission level from the permissions struct
func getPermissionLevel(perms *github.RepositoryPermissions) string {
	if perms.GetAdmin() {
		return "admin"
	}
	if perms.GetMaintain() {
		return "maintain"
	}
	if perms.GetPush() {
		return "write"
	}
	if perms.GetTriage() {
		return "triage"
	}
	if perms.GetPull() {
		return "read"
	}
	return "unknown"
}

// BuildAccessMap aggregates user access across all repositories
func (c *Client) BuildAccessMap(ctx context.Context, org string, repoFilter string, permFilter string, includeArchived bool) ([]UserAccess, error) {
	var repos []*github.Repository
	var err error

	if repoFilter != "" {
		// Fetch single repository
		c.logger.Info("fetching repository", "organization", org, "repository", repoFilter)
		repo, resp, err := c.client.Repositories.Get(ctx, org, repoFilter)
		if err != nil {
			return nil, c.handleRateLimit(resp, fmt.Errorf("failed to fetch repository: %w", err))
		}
		repos = []*github.Repository{repo}
	} else {
		// Fetch all repositories
		repos, err = c.FetchOrgRepositories(ctx, org, includeArchived)
		if err != nil {
			return nil, err
		}
	}

	// Map of username -> UserAccess
	userAccessMap := make(map[string]*UserAccess)

	for i, repo := range repos {
		c.logger.Info("processing repository", "progress", fmt.Sprintf("%d/%d", i+1, len(repos)), "repository", repo.GetName())

		collaborators, permissions, err := c.FetchRepoCollaborators(ctx, org, repo.GetName())
		if err != nil {
			c.logger.Warn("failed to fetch collaborators", "repository", repo.GetName(), "error", err)
			continue
		}

		for _, collab := range collaborators {
			username := collab.GetLogin()
			permission := permissions[username]

			// Filter by permission level if specified
			if permFilter != "" && permission != permFilter {
				continue
			}

			// Skip read-only access (we only want triage, write, maintain, admin)
			if permission == "read" || permission == "unknown" {
				continue
			}

			// Add or update user access
			if userAccess, exists := userAccessMap[username]; exists {
				userAccess.Repos = append(userAccess.Repos, RepoPermission{
					RepoName:   repo.GetName(),
					Permission: permission,
				})
			} else {
				userAccessMap[username] = &UserAccess{
					Username: username,
					Repos: []RepoPermission{
						{
							RepoName:   repo.GetName(),
							Permission: permission,
						},
					},
				}
			}
		}
	}

	// Convert map to slice
	var userAccesses []UserAccess
	for _, ua := range userAccessMap {
		userAccesses = append(userAccesses, *ua)
	}

	return userAccesses, nil
}
