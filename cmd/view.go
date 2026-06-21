package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/jamesjohnsdev/issues/internal/issue"
	"github.com/spf13/cobra"
)

var viewCommentsFlag bool
var viewWebFlag bool

var viewCmd = &cobra.Command{
	Use:   "view <number>",
	Short: "Open an issue in the default editor",
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

		if viewWebFlag {
			if iss.Number == 0 {
				return fmt.Errorf("issue %s is local-only and has no GitHub URL", idFromPath(iss.Path))
			}
			c := exec.Command("gh", "issue", "view", fmt.Sprintf("%d", iss.Number), "--web")
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			return c.Run()
		}

		if viewCommentsFlag {
			return printComments(iss)
		}

		editor := os.Getenv("VISUAL")
		if editor == "" {
			editor = os.Getenv("EDITOR")
		}
		if editor == "" {
			return fmt.Errorf("no editor set: define $VISUAL or $EDITOR")
		}

		parts := strings.Fields(editor)
		c := exec.Command(parts[0], append(parts[1:], iss.Path)...)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	},
}

func printComments(iss *issue.Issue) error {
	commentsPath := filepath.Join(filepath.Dir(iss.Path), issue.CommentsFilename(iss))
	comments, err := issue.ParseComments(commentsPath)
	if err != nil {
		return fmt.Errorf("reading comments: %w", err)
	}

	if len(comments) == 0 {
		fmt.Printf("%s no comments on #%d\n", color.New(color.FgHiBlack).Sprint("—"), iss.Number)
		return nil
	}

	bold := color.New(color.Bold).SprintFunc()
	dim := color.New(color.FgHiBlack).SprintFunc()
	draft := color.New(color.FgYellow).Sprint("[draft]")

	for i, c := range comments {
		if i > 0 {
			fmt.Println()
		}
		if c.Metadata == nil {
			fmt.Printf("%s %s\n", bold("draft"), draft)
		} else {
			ts := ""
			if c.Metadata.CreatedAt != nil {
				ts = " · " + c.Metadata.CreatedAt.In(time.Local).Format("2 Jan 2006 15:04")
			}
			fmt.Printf("%s%s\n", bold(c.Metadata.Author), dim(ts))
		}
		fmt.Println(c.Body)
	}
	return nil
}

func init() {
	viewCmd.Flags().BoolVarP(&viewCommentsFlag, "comments", "c", false, "show comments instead of opening the editor")
	viewCmd.Flags().BoolVarP(&viewWebFlag, "web", "w", false, "open the issue in the browser")
}
