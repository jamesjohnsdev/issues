package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List local issues",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := issuesRoot()
		if err != nil {
			return err
		}
		issues, err := loadAllLocal(root)
		if err != nil {
			return err
		}
		if len(issues) == 0 {
			fmt.Println("No local issues. Run `issue pull` to fetch from GitHub.")
			return nil
		}
		for _, iss := range issues {
			fmt.Printf("%-8s %-8s %s\n", idFromPath(iss.Path), "["+iss.State+"]", iss.Title)
		}
		return nil
	},
}
