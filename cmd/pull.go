package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"

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
			return pullOne(root, remote)
		}

		remotes, err := gh.ListAll()
		if err != nil {
			return err
		}
		for _, remote := range remotes {
			if err := pullOne(root, remote); err != nil {
				return err
			}
		}
		fmt.Printf("Pulled %d issue(s)\n", len(remotes))
		return nil
	},
}

func pullOne(root string, iss *issue.Issue) error {
	now := time.Now().UTC().Truncate(time.Second)
	iss.SyncedAt = &now

	dir := stateDir(root, iss.State)
	destPath := filepath.Join(dir, issue.Filename(iss))

	// If the issue exists locally in a different location (e.g. state changed), remove the old file
	existing, _ := findLocalByNumber(root, iss.Number)
	if existing != nil && existing.Path != destPath {
		os.Remove(existing.Path)
	}

	if err := issue.Write(destPath, iss); err != nil {
		return fmt.Errorf("writing #%d: %w", iss.Number, err)
	}

	origPath := filepath.Join(originalsDir(root), fmt.Sprintf("%d.md", iss.Number))
	return saveOriginal(destPath, origPath)
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
