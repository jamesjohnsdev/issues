package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/jamesjohnsdev/issues/internal/gh"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete an issue locally (and optionally from GitHub)",
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

		if iss.Number != 0 {
			fmt.Printf("Issue #%d (%s) is synced to GitHub.\n", iss.Number, iss.Title)
			fmt.Print("Delete from GitHub as well? [y/N] ")

			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))

			if answer == "y" || answer == "yes" {
				if err := gh.Delete(iss.Number); err != nil {
					return err
				}
				fmt.Printf("%s #%d from GitHub\n", color.RedString("Deleted"), iss.Number)
			}

		}

		if err := os.Remove(iss.Path); err != nil {
			return err
		}

		if iss.Number != 0 {
			origPath := filepath.Join(originalsDir(root), fmt.Sprintf("%d.md", iss.Number))
			_ = os.Remove(origPath)
		}

		fmt.Printf("%s %s: %s\n", color.RedString("Deleted"), args[0], iss.Title)
		return nil
	},
}
