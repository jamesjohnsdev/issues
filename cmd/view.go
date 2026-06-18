package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

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
