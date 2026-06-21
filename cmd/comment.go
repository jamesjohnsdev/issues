package cmd

import (
	"fmt"
	"path/filepath"

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

		body, err := draftCommentBody()
		if err != nil {
			return err
		}
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
