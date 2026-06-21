package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/jamesjohnsdev/issues/internal/gh"
	"github.com/jamesjohnsdev/issues/internal/issue"
	"github.com/spf13/cobra"
)

var mergeCmd = &cobra.Command{
	Use:   "merge <a> <b>",
	Short: "Close issue a as a duplicate of issue b",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := issuesRoot()
		if err != nil {
			return err
		}

		if args[0] == args[1] {
			return fmt.Errorf("cannot merge an issue into itself: %q", args[0])
		}

		issA, err := findLocalByID(root, args[0])
		if err != nil {
			return err
		}
		issB, err := findLocalByID(root, args[1])
		if err != nil {
			return err
		}

		if issA.Path == issB.Path {
			return fmt.Errorf("cannot merge an issue into itself: %q", args[0])
		}

		aOnline := issA.Number != 0
		bOnline := issB.Number != 0

		switch {
		case aOnline && bOnline:
			return mergeBothOnline(root, issA, issB)
		case !aOnline && bOnline:
			return mergeLocalOnline(issA, issB)
		case aOnline && !bOnline:
			return mergeOnlineLocal(root, issA, issB)
		default:
			return mergeBothLocal(root, issA, issB)
		}
	},
}

// mergeBothOnline: both issues are on GitHub.
// Closes a with a "duplicate of b" comment, adds a note to b, updates local state.
func mergeBothOnline(root string, issA, issB *issue.Issue) error {
	if err := gh.Close(issA.Number, fmt.Sprintf("Duplicate of #%d", issB.Number)); err != nil {
		return err
	}
	if err := gh.AddComment(issB.Number, fmt.Sprintf("Closed #%d as a duplicate of this issue.", issA.Number)); err != nil {
		return err
	}

	issA.State = "closed"
	issA.StateReason = "not_planned"
	newPath := filepath.Join(closedDir(root), filepath.Base(issA.Path))
	if err := os.MkdirAll(closedDir(root), 0755); err != nil {
		return err
	}
	if err := issue.Write(newPath, issA); err != nil {
		return err
	}
	if err := os.Remove(issA.Path); err != nil {
		return err
	}
	commentsPath := filepath.Join(filepath.Dir(issA.Path), issue.CommentsFilename(issA))
	if err := os.Rename(commentsPath, filepath.Join(closedDir(root), filepath.Base(commentsPath))); err != nil {
		return err
	}

	fmt.Printf("%s #%d (%s) → #%d (%s)\n",
		color.GreenString("Merged"), issA.Number, issA.Title, issB.Number, issB.Title)
	return nil
}

// mergeLocalOnline: a is local-only, b is on GitHub.
// Adds a's content as a comment on b, then deletes a locally.
func mergeLocalOnline(issA, issB *issue.Issue) error {
	aID := idFromPath(issA.Path)
	body := fmt.Sprintf("Merged local issue %s: **%s**\n\n%s", aID, issA.Title, issA.Body)
	if err := gh.AddComment(issB.Number, body); err != nil {
		return err
	}

	commentsPath := filepath.Join(filepath.Dir(issA.Path), issue.CommentsFilename(issA))
	_ = os.Remove(commentsPath)
	if err := os.Remove(issA.Path); err != nil {
		return err
	}

	fmt.Printf("%s %s (%s) → #%d (%s), deleted locally\n",
		color.GreenString("Merged"), aID, issA.Title, issB.Number, issB.Title)
	return nil
}

// mergeOnlineLocal: a is on GitHub, b is local-only.
// Prompts for confirmation, adds a's content as a local comment on b, deletes a from GitHub.
func mergeOnlineLocal(root string, issA, issB *issue.Issue) error {
	bID := idFromPath(issB.Path)
	fmt.Printf("Issue #%d (%s) is on GitHub and will be deleted from GitHub.\n", issA.Number, issA.Title)
	fmt.Print("Proceed? [y/N] ")

	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	if answer = strings.TrimSpace(strings.ToLower(answer)); answer != "y" && answer != "yes" {
		fmt.Println(color.YellowString("Aborted."))
		return nil
	}

	commentBody := fmt.Sprintf("Merged from GitHub #%d: **%s**\n\n%s", issA.Number, issA.Title, issA.Body)
	commentsPath := filepath.Join(filepath.Dir(issB.Path), issue.CommentsFilename(issB))
	comments, err := issue.ParseComments(commentsPath)
	if err != nil {
		return fmt.Errorf("reading comments: %w", err)
	}
	comments = append(comments, &issue.Comment{Body: commentBody})
	if err := issue.WriteComments(commentsPath, comments); err != nil {
		return fmt.Errorf("saving comment: %w", err)
	}

	if err := gh.Delete(issA.Number); err != nil {
		return err
	}
	origPath := filepath.Join(originalsDir(root), fmt.Sprintf("%d.md", issA.Number))
	_ = os.Remove(origPath)
	if err := os.Remove(issA.Path); err != nil {
		return err
	}

	fmt.Printf("%s #%d (%s) → %s (%s), deleted from GitHub\n",
		color.GreenString("Merged"), issA.Number, issA.Title, bID, issB.Title)
	return nil
}

// mergeBothLocal: both issues are local-only.
// Adds b's reference as a comment on a, then closes a locally.
func mergeBothLocal(root string, issA, issB *issue.Issue) error {
	aID := idFromPath(issA.Path)
	bID := idFromPath(issB.Path)

	commentBody := fmt.Sprintf("Closed as duplicate of %s: **%s**", bID, issB.Title)
	commentsPath := filepath.Join(filepath.Dir(issA.Path), issue.CommentsFilename(issA))
	comments, err := issue.ParseComments(commentsPath)
	if err != nil {
		return fmt.Errorf("reading comments: %w", err)
	}
	comments = append(comments, &issue.Comment{Body: commentBody})
	if err := issue.WriteComments(commentsPath, comments); err != nil {
		return fmt.Errorf("saving comment: %w", err)
	}

	issA.State = "closed"
	newPath := filepath.Join(closedDir(root), filepath.Base(issA.Path))
	if err := os.MkdirAll(closedDir(root), 0755); err != nil {
		return err
	}
	if err := issue.Write(newPath, issA); err != nil {
		return err
	}
	if err := os.Remove(issA.Path); err != nil {
		return err
	}
	newCommentsPath := filepath.Join(closedDir(root), filepath.Base(commentsPath))
	_ = os.Rename(commentsPath, newCommentsPath)

	fmt.Printf("%s %s (%s) → %s (%s)\n",
		color.GreenString("Merged"), aID, issA.Title, bID, issB.Title)
	return nil
}
