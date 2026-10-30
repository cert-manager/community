package report

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/cert-manager/community/security/audit-access/pkg/github"
)

// AccessEntry represents a single user-repo-permission entry
type AccessEntry struct {
	Username   string
	RepoName   string
	Permission string
}

// Format formats and outputs the user access data in the specified format
func Format(userAccesses []github.UserAccess, format, sortBy string) error {
	// Convert to flat list of entries for easier sorting
	var entries []AccessEntry
	for _, ua := range userAccesses {
		for _, repo := range ua.Repos {
			entries = append(entries, AccessEntry{
				Username:   ua.Username,
				RepoName:   repo.RepoName,
				Permission: repo.Permission,
			})
		}
	}

	// Sort based on sortBy parameter
	if sortBy == "repo" {
		// Sort by repository first, then by username
		sort.Slice(entries, func(i, j int) bool {
			if entries[i].RepoName != entries[j].RepoName {
				return entries[i].RepoName < entries[j].RepoName
			}
			return entries[i].Username < entries[j].Username
		})
	} else {
		// Sort by username first, then by repository (default)
		sort.Slice(entries, func(i, j int) bool {
			if entries[i].Username != entries[j].Username {
				return entries[i].Username < entries[j].Username
			}
			return entries[i].RepoName < entries[j].RepoName
		})
	}

	switch strings.ToLower(format) {
	case "json":
		if sortBy == "repo" {
			return FormatJSONByRepo(entries)
		}
		return FormatJSONByUser(entries)
	case "csv":
		return FormatCSV(entries)
	case "table":
		return FormatTable(entries)
	default:
		return fmt.Errorf("unsupported format: %s (supported: table, json, csv)", format)
	}
}

// FormatTable outputs the data as an ASCII table
func FormatTable(entries []AccessEntry) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	// Print header
	fmt.Fprintln(w, "USERNAME\tREPOSITORY\tPERMISSION")
	fmt.Fprintln(w, "--------\t----------\t----------")

	// Print data
	for _, entry := range entries {
		fmt.Fprintf(w, "%s\t%s\t%s\n", entry.Username, entry.RepoName, entry.Permission)
	}

	return w.Flush()
}

// FormatJSONByUser outputs the data as JSON grouped by user
func FormatJSONByUser(entries []AccessEntry) error {
	// Group by user
	userMap := make(map[string][]github.RepoPermission)
	for _, entry := range entries {
		userMap[entry.Username] = append(userMap[entry.Username], github.RepoPermission{
			RepoName:   entry.RepoName,
			Permission: entry.Permission,
		})
	}

	// Convert to slice
	var userAccesses []github.UserAccess
	for username, repos := range userMap {
		userAccesses = append(userAccesses, github.UserAccess{
			Username: username,
			Repos:    repos,
		})
	}

	// Sort by username
	sort.Slice(userAccesses, func(i, j int) bool {
		return userAccesses[i].Username < userAccesses[j].Username
	})

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(userAccesses)
}

// RepoAccess represents a repository and its users
type RepoAccess struct {
	RepoName string         `json:"repository"`
	Users    []UserPermInfo `json:"users"`
}

// UserPermInfo represents a user and their permission
type UserPermInfo struct {
	Username   string `json:"username"`
	Permission string `json:"permission"`
}

// FormatJSONByRepo outputs the data as JSON grouped by repository
func FormatJSONByRepo(entries []AccessEntry) error {
	// Group by repository
	repoMap := make(map[string][]UserPermInfo)
	for _, entry := range entries {
		repoMap[entry.RepoName] = append(repoMap[entry.RepoName], UserPermInfo{
			Username:   entry.Username,
			Permission: entry.Permission,
		})
	}

	// Convert to slice
	var repoAccesses []RepoAccess
	for repoName, users := range repoMap {
		repoAccesses = append(repoAccesses, RepoAccess{
			RepoName: repoName,
			Users:    users,
		})
	}

	// Sort by repository name
	sort.Slice(repoAccesses, func(i, j int) bool {
		return repoAccesses[i].RepoName < repoAccesses[j].RepoName
	})

	// Sort users within each repo
	for i := range repoAccesses {
		sort.Slice(repoAccesses[i].Users, func(a, b int) bool {
			return repoAccesses[i].Users[a].Username < repoAccesses[i].Users[b].Username
		})
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(repoAccesses)
}

// FormatCSV outputs the data as CSV
func FormatCSV(entries []AccessEntry) error {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"Username", "Repository", "Permission"}); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data
	for _, entry := range entries {
		if err := writer.Write([]string{entry.Username, entry.RepoName, entry.Permission}); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}
