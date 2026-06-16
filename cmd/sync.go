package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/jamesjohnsdev/issues/internal/gh"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Push new local issues then pull all from GitHub",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := issuesRoot()
		if err != nil {
			return err
		}

		// Step 1: push any new T-numbered issues
		local, err := loadAllLocal(root)
		if err != nil {
			return err
		}
		newCount := 0
		for _, iss := range local {
			if iss.Number == 0 {
				fmt.Printf("Pushing new issue: %s\n", iss.Title)
				if err := pushOne(root, iss); err != nil {
					return err
				}
				newCount++
			}
		}
		if newCount > 0 {
			fmt.Printf("Pushed %d new issue(s)\n\n", newCount)
		}

		// Reload — T-issues now have real numbers
		local, err = loadAllLocal(root)
		if err != nil {
			return err
		}

		// Step 2: warn about local modifications that will be overwritten
		var modified []string
		for _, iss := range local {
			if iss.Number == 0 {
				continue
			}
			mod, err := isModified(root, iss)
			if err != nil {
				return err
			}
			if mod {
				modified = append(modified, fmt.Sprintf("  #%d: %s", iss.Number, iss.Title))
			}
		}

		if len(modified) > 0 {
			fmt.Println("Warning: the following issues have local changes that will be overwritten by pull:")
			for _, m := range modified {
				fmt.Println(m)
			}
			fmt.Print("\nContinue? [y/N] ")
			line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
			if strings.ToLower(strings.TrimSpace(line)) != "y" {
				fmt.Println("Aborted. Use `issue push` to send your local changes to GitHub first.")
				return nil
			}
			fmt.Println()
		}

		// Step 3: pull all from GitHub
		remotes, err := gh.ListAll()
		if err != nil {
			return err
		}
		for _, remote := range remotes {
			if err := pullOne(root, remote); err != nil {
				return err
			}
		}
		fmt.Printf("Synced %d issue(s) from GitHub\n", len(remotes))
		return nil
	},
}
