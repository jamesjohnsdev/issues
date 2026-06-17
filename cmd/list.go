package cmd

import (
	"fmt"

	"github.com/fatih/color"
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
		idColor := color.New(color.FgCyan)
		openColor := color.New(color.FgGreen)
		closedColor := color.New(color.FgHiBlack)
		for _, iss := range filtered {
			id := idColor.Sprintf("%-8s", idFromPath(iss.Path))
			var state string
			if iss.State == "open" {
				state = openColor.Sprintf("%-8s", "["+iss.State+"]")
			} else {
				state = closedColor.Sprintf("%-8s", "["+iss.State+"]")
			}
			fmt.Printf("%s %s %s\n", id, state, iss.Title)
		}
		return nil
	},
}

func init() {
	listCmd.Flags().BoolVar(&listAll, "all", false, "show open and closed issues")
	listCmd.Flags().BoolVar(&listClosed, "closed", false, "show only closed issues")
}
