package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fatih/color"
	"github.com/jamesjohnsdev/issues/internal/gh"
	"github.com/jamesjohnsdev/issues/internal/issue"
	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull [number]",
	Short: "Pull issue(s) from GitHub",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := issuesRoot()
		if err != nil {
			return err
		}

		if len(args) == 1 {
			var number int
			if _, err := fmt.Sscanf(args[0], "%d", &number); err != nil {
				return fmt.Errorf("invalid issue number: %s", args[0])
			}
			remote, err := gh.Get(number)
			if err != nil {
				return err
			}
			local, _ := findLocalByNumber(root, number)
			localIndex := map[int]*issue.Issue{}
			if local != nil {
				localIndex[local.Number] = local
			}
			commentsPath, err := pullOne(root, remote, localIndex)
			if err != nil {
				return err
			}
			if commentsPath != "" {
				return pullComments(remote, commentsPath)
			}
			return nil
		}

		localIndex, err := buildLocalIndex(root)
		if err != nil {
			return err
		}
		remotes, err := gh.ListAll()
		if err != nil {
			return err
		}
		for _, remote := range remotes {
			commentsPath, err := pullOne(root, remote, localIndex)
			if err != nil {
				return err
			}
			if commentsPath != "" {
				if err := pullComments(remote, commentsPath); err != nil {
					return err
				}
			}
		}
		fmt.Printf("%s %d issue(s)\n", color.GreenString("Pulled"), len(remotes))
		return nil
	},
}

// pullOne writes the issue file and original snapshot. Returns commentsPath to
// be fetched, or empty string if the issue is unchanged since last sync.
func pullOne(root string, iss *issue.Issue, localIndex map[int]*issue.Issue) (string, error) {
	existing := localIndex[iss.Number]

	// Skip if nothing has changed since we last synced.
	if existing != nil && existing.SyncedAt != nil && iss.UpdatedAt != nil && !iss.UpdatedAt.After(*existing.SyncedAt) {
		return "", nil
	}

	now := time.Now().UTC().Truncate(time.Second)
	iss.SyncedAt = &now

	dir := stateDir(root, iss.State)
	destPath := filepath.Join(dir, issue.Filename(iss))
	commentsPath := filepath.Join(dir, issue.CommentsFilename(iss))

	// If state changed the file lives in a different directory; move old files.
	if existing != nil && existing.Path != destPath {
		if err := os.Remove(existing.Path); err != nil {
			return "", fmt.Errorf("removing old local issue: %w", err)
		}
		oldCommentsPath := filepath.Join(filepath.Dir(existing.Path), issue.CommentsFilename(existing))
		if _, err := os.Stat(oldCommentsPath); err == nil {
			if err := os.Rename(oldCommentsPath, commentsPath); err != nil {
				return "", fmt.Errorf("moving comments file: %w", err)
			}
		}
	}

	if err := issue.Write(destPath, iss); err != nil {
		return "", fmt.Errorf("writing #%d: %w", iss.Number, err)
	}

	origPath := filepath.Join(originalsDir(root), fmt.Sprintf("%d.md", iss.Number))
	if err := saveOriginal(destPath, origPath); err != nil {
		return "", err
	}

	return commentsPath, nil
}

func pullComments(iss *issue.Issue, commentsPath string) error {
	remote, err := gh.GetComments(iss.Number)
	if err != nil {
		return err
	}

	// Preserve any local-only comments (those without an id) by merging them in
	local, _ := issue.ParseComments(commentsPath)
	remoteIDs := make(map[string]bool, len(remote))
	for _, c := range remote {
		if c.Metadata != nil {
			remoteIDs[c.Metadata.ID] = true
		}
	}
	for _, c := range local {
		if c.Metadata == nil {
			remote = append(remote, c)
		}
	}
	_ = remoteIDs

	return issue.WriteComments(commentsPath, remote)
}

func saveOriginal(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

func isModified(root string, iss *issue.Issue) (bool, error) {
	origPath := filepath.Join(originalsDir(root), fmt.Sprintf("%d.md", iss.Number))
	orig, err := os.ReadFile(origPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	current, err := os.ReadFile(iss.Path)
	if err != nil {
		return false, err
	}
	return !bytes.Equal(orig, current), nil
}
