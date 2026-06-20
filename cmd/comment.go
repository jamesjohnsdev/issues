package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/jamesjohnsdev/issues/internal/issue"
	"github.com/spf13/cobra"
)

var commentCmd = &cobra.Command{
	Use:   "comment <number>",
	Short: "Draft a new comment on an issue",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := issuesRoot()
		if err != nil {
			return err
		}

		iss, err := findLocalByID(root, args[0])
		if err != nil {
			return err
		}

		editor := os.Getenv("VISUAL")
		if editor == "" {
			editor = os.Getenv("EDITOR")
		}
		if editor == "" {
			return fmt.Errorf("no editor set: define $VISUAL or $EDITOR")
		}

		tmp, err := os.CreateTemp("", "issues-comment-*.md")
		if err != nil {
			return fmt.Errorf("creating temp file: %w", err)
		}
		tmpPath := tmp.Name()
		defer os.Remove(tmpPath)
		tmp.Close()

		parts := strings.Fields(editor)
		c := exec.Command(parts[0], append(parts[1:], tmpPath)...)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			return fmt.Errorf("editor exited with error: %w", err)
		}

		data, err := os.ReadFile(tmpPath)
		if err != nil {
			return fmt.Errorf("reading temp file: %w", err)
		}
		body := strings.TrimSpace(string(data))
		if body == "" {
			fmt.Println(color.YellowString("Aborted.") + " Empty body, comment discarded.")
			return nil
		}

		commentsPath := filepath.Join(filepath.Dir(iss.Path), issue.CommentsFilename(iss))
		comments, err := issue.ParseComments(commentsPath)
		if err != nil {
			return fmt.Errorf("reading comments: %w", err)
		}
		comments = append(comments, &issue.Comment{Body: body})
		if err := issue.WriteComments(commentsPath, comments); err != nil {
			return fmt.Errorf("saving comment: %w", err)
		}

		fmt.Printf("%s comment draft on #%d — run %s to send\n",
			color.GreenString("Saved"),
			iss.Number,
			color.CyanString("issues push %d", iss.Number),
		)
		return nil
	},
}
