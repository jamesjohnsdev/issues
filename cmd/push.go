package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fatih/color"
	"github.com/jamesjohnsdev/issues/internal/gh"
	"github.com/jamesjohnsdev/issues/internal/issue"
	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push [number]",
	Short: "Push local changes to GitHub (all modified, or a specific issue)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := issuesRoot()
		if err != nil {
			return err
		}

		if len(args) == 1 {
			iss, err := findLocalByID(root, args[0])
			if err != nil {
				return err
			}
			return pushOne(root, iss)
		}

		// Push all: create new T-issues, update locally-modified existing ones
		issues, err := loadAllLocal(root)
		if err != nil {
			return err
		}
		pushed := 0
		for _, iss := range issues {
			if iss.Number == 0 {
				if err := pushOne(root, iss); err != nil {
					return err
				}
				pushed++
				continue
			}
			mod, err := isModified(root, iss)
			if err != nil {
				return err
			}
			if mod {
				if err := pushOne(root, iss); err != nil {
					return err
				}
				pushed++
			}
		}
		if pushed == 0 {
			if _, err := color.New(color.FgHiBlack).Println("Nothing to push."); err != nil {
				return err
			}
		} else {
			fmt.Printf("%s %d issue(s)\n", color.GreenString("Pushed"), pushed)
		}
		return nil
	},
}

func pushOne(root string, iss *issue.Issue) error {
	if iss.Number == 0 {
		number, err := gh.Create(iss)
		if err != nil {
			return err
		}
		oldPath := iss.Path
		iss.Number = number
		newPath := filepath.Join(filepath.Dir(oldPath), issue.Filename(iss))
		if err := os.Rename(oldPath, newPath); err != nil {
			return err
		}
		iss.Path = newPath
		// Pull back the canonical GH version
		remote, err := gh.Get(number)
		if err != nil {
			return err
		}
		fmt.Printf("%s #%d: %s\n", color.GreenString("Created"), number, iss.Title)
		return pullOne(root, remote)
	}

	if err := gh.Update(iss); err != nil {
		return err
	}

	// Update synced_at locally and snapshot as new original
	now := time.Now().UTC().Truncate(time.Second)
	iss.SyncedAt = &now
	if err := issue.Write(iss.Path, iss); err != nil {
		return err
	}
	origPath := filepath.Join(originalsDir(root), fmt.Sprintf("%d.md", iss.Number))
	if err := saveOriginal(iss.Path, origPath); err != nil {
		return err
	}

	fmt.Printf("%s #%d: %s\n", color.GreenString("Pushed"), iss.Number, iss.Title)
	return nil
}
