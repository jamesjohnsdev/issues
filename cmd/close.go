package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/jamesjohnsdev/issues/internal/issue"
	"github.com/spf13/cobra"
)

var closeCmd = &cobra.Command{
	Use:   "close <number>",
	Short: "Mark an issue as closed",
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

		if iss.State == "closed" {
			fmt.Printf("Issue %s is already closed.\n", args[0])
			return nil
		}

		addComment, _ := cmd.Flags().GetBool("comment")
		if addComment {
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
		}

		iss.State = "closed"

		newPath := filepath.Join(closedDir(root), filepath.Base(iss.Path))
		if err := os.MkdirAll(closedDir(root), 0755); err != nil {
			return err
		}
		if err := issue.Write(newPath, iss); err != nil {
			return err
		}
		if err := os.Remove(iss.Path); err != nil {
			return err
		}

		oldComments := filepath.Join(filepath.Dir(iss.Path), issue.CommentsFilename(iss))
		if _, statErr := os.Stat(oldComments); statErr == nil {
			newComments := filepath.Join(closedDir(root), issue.CommentsFilename(iss))
			if err := os.Rename(oldComments, newComments); err != nil {
				return err
			}
		}

		fmt.Printf("%s %s: %s\n", color.GreenString("Closed"), args[0], iss.Title)
		return nil
	},
}

func init() {
	closeCmd.Flags().BoolP("comment", "c", false, "open editor to draft a closing comment")
}
