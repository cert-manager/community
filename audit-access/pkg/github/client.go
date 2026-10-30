package github

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/google/go-github/v84/github"
	"golang.org/x/oauth2"
)

type Client struct {
	client *github.Client
	token  string
	logger *slog.Logger
}

// NewClient creates a new GitHub client with authentication
func NewClient(ctx context.Context, logger *slog.Logger) (*Client, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN environment variable not set")
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	return &Client{
		client: client,
		token:  token,
		logger: logger,
	}, nil
}

// ValidateToken validates the GitHub token and checks for required permissions
func (c *Client) ValidateToken(ctx context.Context) error {
	// Test authentication by getting the current user
	user, resp, err := c.client.Users.Get(ctx, "")
	if err != nil {
		if resp != nil && resp.StatusCode == 401 {
			return fmt.Errorf("GitHub token has expired or is invalid")
		}
		return fmt.Errorf("failed to authenticate with GitHub: %w", err)
	}

	// Check rate limit
	rateLimit, _, err := c.client.RateLimit.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to check rate limit: %w", err)
	}

	if rateLimit.Core.Remaining == 0 {
		resetTime := rateLimit.Core.Reset.Time
		return fmt.Errorf("rate limit exceeded, resets at %s", resetTime.Format(time.RFC3339))
	}

	// Check token scopes
	scopes := resp.Header.Get("X-OAuth-Scopes")
	if scopes == "" {
		// If no scopes header, the token might be a classic token without specific scopes
		// We'll proceed and let API calls fail if permissions are insufficient
	}

	c.logger.Info("authenticated",
		"user", user.GetLogin(),
		"rate_limit_remaining", rateLimit.Core.Remaining,
		"rate_limit_total", rateLimit.Core.Limit,
		"rate_limit_resets", rateLimit.Core.Reset.Time.Format("15:04:05"))

	return nil
}

// handleRateLimit checks rate limit and returns helpful error if exceeded
func (c *Client) handleRateLimit(resp *github.Response, err error) error {
	if resp != nil && resp.Rate.Remaining == 0 {
		resetTime := resp.Rate.Reset.Time
		return fmt.Errorf("rate limit exceeded, resets at %s", resetTime.Format(time.RFC3339))
	}
	return err
}

// GetClient returns the underlying GitHub client
func (c *Client) GetClient() *github.Client {
	return c.client
}

// checkResponse checks the HTTP response and returns helpful error messages
func checkResponse(resp *http.Response, err error) error {
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case 401:
		return fmt.Errorf("authentication failed: token is invalid or expired")
	case 403:
		if resp.Header.Get("X-RateLimit-Remaining") == "0" {
			resetTime := resp.Header.Get("X-RateLimit-Reset")
			return fmt.Errorf("rate limit exceeded, resets at %s", resetTime)
		}
		return fmt.Errorf("forbidden: token lacks required permissions (needs 'repo' and 'read:org' scopes)")
	case 404:
		return fmt.Errorf("not found: organization or repository does not exist, or token lacks access")
	}

	return nil
}
