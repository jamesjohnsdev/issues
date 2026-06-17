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

		fmt.Printf("%s %s: %s\n", color.GreenString("Closed"), args[0], iss.Title)
		return nil
	},
}
