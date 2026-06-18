package cmd

import (
	"os"
	"path/filepath"
	"strings"
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
	f.Cleanup(func() {
		if err := os.RemoveAll(root); err != nil {
			f.Fatal(err)
		}
	})

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

func FuzzCreateInteractive(f *testing.F) {
	f.Add("Fix the bug", "", "", "", "")
	f.Add("Add feature", "e", "bug,feature", "alice", "v2.0")
	f.Add("", "", "", "", "")
	f.Add("title: with colon", "", "l1, l2", "a, b", "ms 1")
	f.Add(strings.Repeat("x", 200), "", "label", "", "")
	f.Add("  ", "", "", "", "")
	f.Add("Nö ASCII", "", "", "", "")
	f.Add("a", "e", "", "", "")

	f.Fuzz(func(t *testing.T, title, bodyResp, labels, assignees, milestone string) {
		parent := makeProjectDir(t, nil)
		chdirTo(t, parent)

		orig := createEditorFlag
		t.Cleanup(func() { createEditorFlag = orig })
		createEditorFlag = false

		t.Setenv("VISUAL", "")
		t.Setenv("EDITOR", "")

		// Strip embedded newlines from each field so they don't bleed into
		// adjacent prompts.
		clean := func(s string) string { return strings.ReplaceAll(s, "\n", " ") }
		stdin := strings.Join([]string{
			clean(title), clean(bodyResp), clean(labels), clean(assignees), clean(milestone),
		}, "\n") + "\n"

		injectStdin(t, stdin)
		_ = captureStdout(t, func() {
			_ = createCmd.RunE(createCmd, nil)
		})

		openDir := filepath.Join(parent, issuesDirName, "open")
		files := readMDFiles(t, openDir)

		wantTitle := strings.TrimSpace(title)
		if wantTitle == "" {
			if len(files) != 0 {
				t.Errorf("empty title produced %d files, want 0", len(files))
			}
			return
		}

		if len(files) != 1 {
			t.Fatalf("title %q produced %d files, want 1", title, len(files))
		}

		iss, err := issue.Parse(files[0])
		if err != nil {
			t.Fatalf("created file not parseable: %v", err)
		}
		if iss.Title != wantTitle {
			t.Errorf("Title = %q, want %q", iss.Title, wantTitle)
		}
		if iss.State != "open" {
			t.Errorf("State = %q, want open", iss.State)
		}
	})
}
