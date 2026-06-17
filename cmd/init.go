package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a .issues directory in the current directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		root := filepath.Join(cwd, issuesDirName)
		for _, d := range []string{
			filepath.Join(root, "open"),
			filepath.Join(root, "closed"),
			filepath.Join(root, ".sync", "originals"),
		} {
			if err := os.MkdirAll(d, 0755); err != nil {
				return err
			}
		}
		fmt.Printf("%s %s\n", color.GreenString("Initialized"), root)
		return nil
	},
}
