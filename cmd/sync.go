package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/jamesjohnsdev/issues/internal/gh"
	"github.com/jamesjohnsdev/issues/internal/issue"
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

		// Step 1: push any new T-numbered issues in parallel
		local, err := loadAllLocal(root)
		if err != nil {
			return err
		}

		var newIssues []*issue.Issue
		for _, iss := range local {
			if iss.Number == 0 {
				newIssues = append(newIssues, iss)
			}
		}

		if len(newIssues) > 0 {
			const concurrency = 10
			sem := make(chan struct{}, concurrency)
			var wg sync.WaitGroup
			errCh := make(chan error, len(newIssues))

			for _, iss := range newIssues {
				wg.Add(1)
				sem <- struct{}{}
				go func(iss *issue.Issue) {
					defer wg.Done()
					defer func() { <-sem }()
					fmt.Printf("Pushing new issue: %s\n", color.CyanString(iss.Title))
					if err := pushOne(root, iss); err != nil {
						errCh <- err
					}
				}(iss)
			}
			wg.Wait()
			close(errCh)
			if err := <-errCh; err != nil {
				return err
			}
			fmt.Printf("%s %d new issue(s)\n\n", color.GreenString("Pushed"), len(newIssues))
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
			fmt.Printf("%s the following issues have local changes that will be overwritten by pull:\n", color.YellowString("Warning:"))
			for _, m := range modified {
				fmt.Println(color.YellowString(m))
			}
			fmt.Print("\nContinue? [y/N] ")
			line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
			if strings.ToLower(strings.TrimSpace(line)) != "y" {
				fmt.Println(color.YellowString("Aborted.") + " Use `issues push` to send your local changes to GitHub first.")
				return nil
			}
			fmt.Println()
		}

		// Step 3: pull all from GitHub
		remotes, err := gh.ListAll()
		if err != nil {
			return err
		}

		// Reload local index after any pushes above.
		localIndex, err := buildLocalIndex(root)
		if err != nil {
			return err
		}

		// Write issue files serially, collect which ones need comment fetches.
		type commentJob struct {
			iss          *issue.Issue
			commentsPath string
		}
		var jobs []commentJob
		for _, remote := range remotes {
			cp, err := pullOne(root, remote, localIndex)
			if err != nil {
				return err
			}
			if cp != "" {
				jobs = append(jobs, commentJob{remote, cp})
			}
		}

		// Fan out comment fetches — one gh subprocess per issue but now concurrent.
		const concurrency = 20
		sem := make(chan struct{}, concurrency)
		var wg sync.WaitGroup
		errCh := make(chan error, len(jobs))
		for _, j := range jobs {
			wg.Add(1)
			sem <- struct{}{}
			go func(j commentJob) {
				defer wg.Done()
				defer func() { <-sem }()
				if err := pullComments(j.iss, j.commentsPath); err != nil {
					errCh <- err
				}
			}(j)
		}
		wg.Wait()
		close(errCh)
		if err := <-errCh; err != nil {
			return err
		}

		skipped := len(remotes) - len(jobs)
		if skipped > 0 {
			fmt.Printf("%s %d issue(s) from GitHub (%d unchanged)\n", color.GreenString("Synced"), len(remotes), skipped)
		} else {
			fmt.Printf("%s %d issue(s) from GitHub\n", color.GreenString("Synced"), len(remotes))
		}
		return nil
	},
}
