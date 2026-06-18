package cmd

import (
	"fmt"
	"sort"
	"strings"

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
		sort.Slice(filtered, func(i, j int) bool {
			ni, nj := filtered[i].Number, filtered[j].Number
			// GitHub issues (Number > 0) sort before local T-issues (Number == 0)
			if (ni == 0) != (nj == 0) {
				return ni != 0
			}
			if ni != 0 {
				return ni < nj
			}
			// Both local: compare T-numbers
			var ti, tj int
			if _, err := fmt.Sscanf(idFromPath(filtered[i].Path), "T%d", &ti); err != nil {
				return false
			}
			if _, err := fmt.Sscanf(idFromPath(filtered[j].Path), "T%d", &tj); err != nil {
				return false
			}
			return ti < tj
		})

		if len(filtered) == 0 {
			fmt.Println("No local issues. Run `issues pull` to fetch from GitHub.")
			return nil
		}

		bold := color.New(color.Bold).SprintfFunc()
		idColor := color.New(color.FgCyan).SprintfFunc()
		localIDColor := color.New(color.FgYellow).SprintfFunc()
		openColor := color.New(color.FgGreen).SprintfFunc()
		closedColor := color.New(color.FgHiBlack).SprintfFunc()

		// Calculate column widths from plain strings, before any color codes are applied
		idW, stateW := len("ID"), len("State")
		for _, iss := range filtered {
			if n := len(idFromPath(iss.Path)); n > idW {
				idW = n
			}
			if n := len(iss.State); n > stateW {
				stateW = n
			}
		}

		const pad = "  "

		fmt.Println()
		fmt.Printf("%s%s  %s  %s\n", pad,
			bold("%-*s", idW, "ID"),
			bold("%-*s", stateW, "State"),
			bold("Title"),
		)
		fmt.Printf("%s%s  %s  %s\n", pad,
			strings.Repeat("─", idW),
			strings.Repeat("─", stateW),
			strings.Repeat("─", 33),
		)
		for _, iss := range filtered {
			var id string
			if iss.Number == 0 {
				id = localIDColor("%-*s", idW, idFromPath(iss.Path))
			} else {
				id = idColor("%-*s", idW, idFromPath(iss.Path))
			}
			var state string
			if iss.State == "open" {
				state = openColor("%-*s", stateW, iss.State)
			} else {
				state = closedColor("%-*s", stateW, iss.State)
			}
			fmt.Printf("%s%s  %s  %s\n", pad, id, state, iss.Title)
		}
		return nil
	},
}

func init() {
	listCmd.Flags().BoolVar(&listAll, "all", false, "show open and closed issues")
	listCmd.Flags().BoolVar(&listClosed, "closed", false, "show only closed issues")
}
