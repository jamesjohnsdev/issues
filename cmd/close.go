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
		var number int
		if _, err := fmt.Sscanf(args[0], "%d", &number); err != nil {
			return fmt.Errorf("invalid issue number: %s", args[0])
		}

		root, err := issuesRoot()
		if err != nil {
			return err
		}

		iss, err := findLocalByNumber(root, number)
		if err != nil {
			return err
		}

		if iss.State == "closed" {
			fmt.Printf("Issue #%d is already closed.\n", number)
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

		fmt.Printf("%s #%d: %s\n", color.GreenString("Closed"), number, iss.Title)
		return nil
	},
}
