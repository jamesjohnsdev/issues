package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var viewCmd = &cobra.Command{
	Use:   "view <number>",
	Short: "Open an issue in the default editor",
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

		editor := os.Getenv("VISUAL")
		if editor == "" {
			editor = os.Getenv("EDITOR")
		}
		if editor == "" {
			return fmt.Errorf("no editor set: define $VISUAL or $EDITOR")
		}

		c := exec.Command(editor, iss.Path)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	},
}
