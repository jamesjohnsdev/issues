package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jamesjohnsdev/issues/internal/issue"
)

const issuesDirName = ".issues"

func issuesRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		path := filepath.Join(cwd, issuesDirName)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
		parentDir := filepath.Dir(cwd)
		if parentDir == cwd {
			break
		}
		cwd = parentDir
	}
	return "", errors.New("no .issues directory — run `issues init` first")
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

func buildLocalIndex(root string) (map[int]*issue.Issue, error) {
	issues, err := loadAllLocal(root)
	if err != nil {
		return nil, err
	}
	m := make(map[int]*issue.Issue, len(issues))
	for _, iss := range issues {
		m[iss.Number] = iss
	}
	return m, nil
}

// findLocalByID accepts either a plain integer ("42") or a T-prefixed local ID ("T1").
func findLocalByID(root, id string) (*issue.Issue, error) {
	issues, err := loadAllLocal(root)
	if err != nil {
		return nil, err
	}

	// T-prefixed local issue
	if strings.HasPrefix(strings.ToUpper(id), "T") {
		for _, iss := range issues {
			if strings.EqualFold(idFromPath(iss.Path), id) {
				return iss, nil
			}
		}
		return nil, fmt.Errorf("issue %s not found locally", id)
	}

	// GitHub issue number
	number, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("invalid issue id: %s", id)
	}
	for _, iss := range issues {
		if iss.Number == number {
			return iss, nil
		}
	}
	return nil, fmt.Errorf("issue #%d not found locally", number)
}

// draftCommentBody opens $VISUAL or $EDITOR for the user to write a comment and returns the trimmed body.
// An empty string means the user saved without writing anything.
func draftCommentBody() (string, error) {
	editor := os.Getenv("VISUAL")
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		return "", fmt.Errorf("no editor set: define $VISUAL or $EDITOR")
	}

	tmp, err := os.CreateTemp("", "issues-comment-*.md")
	if err != nil {
		return "", fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmp.Name()
	var funcErr error
	defer func() {
		if err := os.Remove(tmpPath); err != nil {
			funcErr = fmt.Errorf("removing temp file: %w", err)
		}
	}()
	if funcErr != nil {
		return "", funcErr
	}
	if err := tmp.Close(); err != nil {
		return "", fmt.Errorf("closing temp file: %w", err)
	}

	parts := strings.Fields(editor)
	c := exec.Command(parts[0], append(parts[1:], tmpPath)...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return "", fmt.Errorf("editor exited with error: %w", err)
	}

	data, err := os.ReadFile(tmpPath)
	if err != nil {
		return "", fmt.Errorf("reading temp file: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}

// idFromPath extracts the ID prefix ("T1" or "42") from a filename like "T1-my-issue.md".
func idFromPath(path string) string {
	base := strings.TrimSuffix(filepath.Base(path), ".md")
	parts := strings.SplitN(base, "-", 2)
	return parts[0]
}
