package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	ghclient "github.com/cert-manager/community/security/audit-access/pkg/github"
	"github.com/cert-manager/community/security/audit-access/pkg/report"
	"github.com/spf13/cobra"
)

var (
	format           string
	permissionFilter string
	repoFilter       string
	includeArchived  bool
	sortBy           string
)

func Execute(version string, log *slog.Logger) error {
	rootCmd := &cobra.Command{
		Use:   "audit-access",
		Short: "Audit GitHub organization access permissions",
		Long: `A CLI tool to audit GitHub organization membership and repository access.
Lists all users with triage, write, maintain, or admin permissions across repositories.`,
		Version: version,
	}

	auditCmd := &cobra.Command{
		Use:   "org [organization-name]",
		Short: "Audit access permissions for a GitHub organization",
		Long: `Audits all repositories in a GitHub organization and lists users with elevated permissions.

Requires GITHUB_TOKEN environment variable with 'repo' and 'read:org' scopes.`,
		Args: cobra.ExactArgs(1),
		RunE: runAudit(log),
	}

	auditCmd.Flags().StringVarP(&format, "format", "f", "table", "Output format (table, json, csv)")
	auditCmd.Flags().StringVarP(&permissionFilter, "permission", "p", "", "Filter by permission level (triage, write, maintain, admin)")
	auditCmd.Flags().StringVarP(&repoFilter, "repo", "r", "", "Audit specific repository instead of all")
	auditCmd.Flags().BoolVarP(&includeArchived, "include-archived", "a", false, "Include archived repositories")
	auditCmd.Flags().StringVarP(&sortBy, "sort-by", "s", "user", "Sort output by 'user' or 'repo'")

	rootCmd.AddCommand(auditCmd)
	return rootCmd.Execute()
}

func runAudit(logger *slog.Logger) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		org := args[0]

		// Validate permission filter if provided
		if permissionFilter != "" {
			validPerms := map[string]bool{
				"triage":   true,
				"write":    true,
				"maintain": true,
				"admin":    true,
			}
			if !validPerms[permissionFilter] {
				return fmt.Errorf("invalid permission filter: %s (valid: triage, write, maintain, admin)", permissionFilter)
			}
		}

		// Validate sort-by option
		if sortBy != "user" && sortBy != "repo" {
			return fmt.Errorf("invalid sort-by option: %s (valid: user, repo)", sortBy)
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Minute)
		defer cancel()

		// Create GitHub client
		logger.Info("initializing GitHub client")
		client, err := ghclient.NewClient(ctx, logger)
		if err != nil {
			return err
		}

		// Validate token
		if err := client.ValidateToken(ctx); err != nil {
			return err
		}

		// Build access map
		userAccesses, err := client.BuildAccessMap(ctx, org, repoFilter, permissionFilter, includeArchived)
		if err != nil {
			return err
		}

		if len(userAccesses) == 0 {
			logger.Info("no users found with elevated permissions")
			return nil
		}

		logger.Info("found users with elevated permissions", "count", len(userAccesses))

		// Format and output results
		return report.Format(userAccesses, format, sortBy)
	}
}
