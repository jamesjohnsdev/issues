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

var createEditorFlag bool

var createCmd = &cobra.Command{
	Use:   "create <title>",
	Short: "Create a new local issue",
	Args: func(cmd *cobra.Command, args []string) error {
		if createEditorFlag {
			return nil
		}
		return cobra.ExactArgs(1)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := issuesRoot()
		if err != nil {
			return err
		}

		existing, err := loadAllLocal(root)
		if err != nil {
			return err
		}
		maxT := 0
		for _, e := range existing {
			if e.Number == 0 {
				var n int
				if _, err := fmt.Sscanf(idFromPath(e.Path), "T%d", &n); err != nil {
					return err
				}
				if n > maxT {
					maxT = n
				}
			}
		}

		var title, filename string
		if createEditorFlag {
			title = ""
			filename = fmt.Sprintf("T%d-new-issue.md", maxT+1)
		} else {
			title = args[0]
			filename = fmt.Sprintf("T%d-%s.md", maxT+1, issue.Slug(title))
		}

		iss := &issue.Issue{Title: title, State: "open"}
		path := filepath.Join(openDir(root), filename)

		if err := issue.Write(path, iss); err != nil {
			return err
		}

		editor := os.Getenv("VISUAL")
		if editor == "" {
			editor = os.Getenv("EDITOR")
		}
		if createEditorFlag && editor == "" {
			return fmt.Errorf("no editor set: define $VISUAL or $EDITOR")
		}
		if editor != "" {
			c := exec.Command(editor, path)
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			_ = c.Run()
		}

		if createEditorFlag {
			saved, err := issue.Parse(path)
			if err != nil || saved.Title == "" {
				_ = os.Remove(path)
				fmt.Println(color.YellowString("Aborted.") + " No title set, issue discarded.")
				return nil
			}
		}

		fmt.Printf("%s %s\n", color.GreenString("Created"), path)
		return nil
	},
}

func init() {
	createCmd.Flags().BoolVarP(&createEditorFlag, "editor", "e", false, "open a new blank issue directly in the editor")
}
