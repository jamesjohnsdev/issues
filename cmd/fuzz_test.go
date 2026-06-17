package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jamesjohnsdev/issues/internal/issue"
)

func FuzzFindLocalByID(f *testing.F) {
	f.Add("1")
	f.Add("10")
	f.Add("T1")
	f.Add("t1")
	f.Add("T")
	f.Add("")
	f.Add("abc")
	f.Add("0")
	f.Add("99999")
	f.Add("T99999")

	// Build a shared root once; all fuzz iterations reuse it.
	root, err := os.MkdirTemp("", "fuzz-issues-*")
	if err != nil {
		f.Fatal(err)
	}
	f.Cleanup(func() { os.RemoveAll(root) })

	for _, dir := range []string{
		filepath.Join(root, "open"),
		filepath.Join(root, "closed"),
	} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			f.Fatal(err)
		}
	}

	fixtures := []struct {
		filename string
		iss      issue.Issue
	}{
		{"1-issue.md", issue.Issue{Number: 1, Title: "Issue one", State: "open"}},
		{"10-tenth.md", issue.Issue{Number: 10, Title: "Tenth issue", State: "open"}},
		{"2-closed.md", issue.Issue{Number: 2, Title: "Closed issue", State: "closed"}},
		{"T1-draft.md", issue.Issue{Title: "Local draft", State: "open"}},
	}
	for _, fix := range fixtures {
		dir := filepath.Join(root, "open")
		if fix.iss.State == "closed" {
			dir = filepath.Join(root, "closed")
		}
		iss := fix.iss
		if err := issue.Write(filepath.Join(dir, fix.filename), &iss); err != nil {
			f.Fatal(err)
		}
	}

	f.Fuzz(func(t *testing.T, id string) {
		// Must never panic regardless of input.
		iss, err := findLocalByID(root, id)
		if err != nil {
			return
		}
		// On success, the returned issue must have a non-empty path.
		if iss.Path == "" {
			t.Errorf("findLocalByID(%q) returned issue with empty Path", id)
		}
	})
}
