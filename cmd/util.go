package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesjohnsdev/issues/internal/issue"
)

const issuesDirName = ".issues"

func issuesRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	path := filepath.Join(cwd, issuesDirName)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", errors.New("no .issues directory — run `issue init` first")
	}
	return path, nil
}

func openDir(root string) string      { return filepath.Join(root, "open") }
func closedDir(root string) string    { return filepath.Join(root, "closed") }
func originalsDir(root string) string { return filepath.Join(root, ".sync", "originals") }

func stateDir(root, state string) string {
	if state == "closed" {
		return closedDir(root)
	}
	return openDir(root)
}

func loadAllLocal(root string) ([]*issue.Issue, error) {
	var issues []*issue.Issue
	for _, dir := range []string{openDir(root), closedDir(root)} {
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		for _, e := range entries {
			if e.IsDir() || filepath.Ext(e.Name()) != ".md" {
				continue
			}
			iss, err := issue.Parse(filepath.Join(dir, e.Name()))
			if err != nil {
				return nil, fmt.Errorf("parsing %s: %w", e.Name(), err)
			}
			issues = append(issues, iss)
		}
	}
	return issues, nil
}

func findLocalByNumber(root string, number int) (*issue.Issue, error) {
	issues, err := loadAllLocal(root)
	if err != nil {
		return nil, err
	}
	for _, iss := range issues {
		if iss.Number == number {
			return iss, nil
		}
	}
	return nil, fmt.Errorf("issue #%d not found locally", number)
}

// idFromPath extracts the ID prefix ("T1" or "42") from a filename like "T1-my-issue.md".
func idFromPath(path string) string {
	base := strings.TrimSuffix(filepath.Base(path), ".md")
	parts := strings.SplitN(base, "-", 2)
	return parts[0]
}
