package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/jamesjohnsdev/issues/internal/issue"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create <title>",
	Short: "Create a new local issue",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := issuesRoot()
		if err != nil {
			return err
		}

		// Find highest existing T-number
		existing, err := loadAllLocal(root)
		if err != nil {
			return err
		}
		maxT := 0
		for _, e := range existing {
			if e.Number == 0 {
				var n int
				fmt.Sscanf(idFromPath(e.Path), "T%d", &n)
				if n > maxT {
					maxT = n
				}
			}
		}

		title := args[0]
		iss := &issue.Issue{Title: title, State: "open"}
		filename := fmt.Sprintf("T%d-%s.md", maxT+1, issue.Slug(title))
		path := filepath.Join(openDir(root), filename)

		if err := issue.Write(path, iss); err != nil {
			return err
		}

		editor := os.Getenv("VISUAL")
		if editor == "" {
			editor = os.Getenv("EDITOR")
		}
		if editor != "" {
			c := exec.Command(editor, path)
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			_ = c.Run()
		}

		fmt.Printf("%s %s\n", color.GreenString("Created"), path)
		return nil
	},
}
