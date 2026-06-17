package cmd

import (
	"fmt"

	"github.com/jamesjohnsdev/issues/internal/issue"
	"github.com/spf13/cobra"
)

var listAll, listClosed bool

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
		var filtered []*issue.Issue
		for _, iss := range issues {
			if listAll || (listClosed && iss.State == "closed") || (!listClosed && iss.State == "open") {
				filtered = append(filtered, iss)
			}
		}
		if len(filtered) == 0 {
			fmt.Println("No local issues. Run `issue pull` to fetch from GitHub.")
			return nil
		}
		for _, iss := range filtered {
			fmt.Printf("%-8s %-8s %s\n", idFromPath(iss.Path), "["+iss.State+"]", iss.Title)
		}
		return nil
	},
}

func init() {
	listCmd.Flags().BoolVar(&listAll, "all", false, "show open and closed issues")
	listCmd.Flags().BoolVar(&listClosed, "closed", false, "show only closed issues")
}
