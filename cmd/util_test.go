package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/jamesjohnsdev/issues/internal/issue"
)

type issueFixture struct {
	filename string
	iss      issue.Issue
}

// makeIssuesRoot creates a temp .issues root with open/ and closed/ dirs
// populated from the given fixtures.
func makeIssuesRoot(t *testing.T, fixtures []issueFixture) string {
	t.Helper()
	root := t.TempDir()
	for _, dir := range []string{
		filepath.Join(root, "open"),
		filepath.Join(root, "closed"),
	} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}
	for _, f := range fixtures {
		dir := filepath.Join(root, "open")
		if f.iss.State == "closed" {
			dir = filepath.Join(root, "closed")
		}
		iss := f.iss
		if err := issue.Write(filepath.Join(dir, f.filename), &iss); err != nil {
			t.Fatalf("writing fixture %s: %v", f.filename, err)
		}
	}
	return root
}

// makeProjectDir creates a temp project root containing a .issues dir with
// open/ and closed/ subdirs populated from fixtures. Returns the project root
// (suitable for os.Chdir so that issuesRoot() can find .issues).
func makeProjectDir(t *testing.T, fixtures []issueFixture) string {
	t.Helper()
	parent := t.TempDir()
	issuesDir := filepath.Join(parent, issuesDirName)
	for _, dir := range []string{
		filepath.Join(issuesDir, "open"),
		filepath.Join(issuesDir, "closed"),
	} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}
	for _, f := range fixtures {
		dir := filepath.Join(issuesDir, "open")
		if f.iss.State == "closed" {
			dir = filepath.Join(issuesDir, "closed")
		}
		iss := f.iss
		if err := issue.Write(filepath.Join(dir, f.filename), &iss); err != nil {
			t.Fatalf("writing fixture %s: %v", f.filename, err)
		}
	}
	return parent
}

// chdirTo changes the working directory to dir and restores the original on
// test cleanup.
func chdirTo(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })
}

// injectStdin replaces os.Stdin with a reader containing content for the
// duration of the test.
func injectStdin(t *testing.T, content string) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.WriteString(content); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	old := os.Stdin
	os.Stdin = r
	t.Cleanup(func() {
		os.Stdin = old
		if err := r.Close(); err != nil {
			t.Error(err)
		}
	})
}

// readMDFiles returns absolute paths to all .md files directly inside dir.
func readMDFiles(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".md" {
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}
	return files
}

// captureStdout runs fn and returns everything written to os.Stdout during its
// execution.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w
	fn()
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdout = old
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func TestDraftCommentBodyWhitespaceEditor(t *testing.T) {
	t.Run("whitespace-only EDITOR returns error instead of panic", func(t *testing.T) {
		t.Setenv("VISUAL", "")
		t.Setenv("EDITOR", "   ")

		_, err := draftCommentBody()
		if err == nil {
			t.Error("expected error for whitespace-only EDITOR, got nil")
		}
	})
}

func TestIDFromPath(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/some/dir/1-hello-world.md", "1"},
		{"/some/dir/T1-local-issue.md", "T1"},
		{"/some/dir/10-double-digit.md", "10"},
		{"42-bare-filename.md", "42"},
		{"/some/dir/T10-local-double.md", "T10"},
		{"/some/dir/nodash.md", "nodash"},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := idFromPath(tt.path)
			if got != tt.want {
				t.Errorf("idFromPath(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestLoadAllLocal(t *testing.T) {
	t.Run("empty dirs returns no issues", func(t *testing.T) {
		root := makeIssuesRoot(t, nil)
		issues, err := loadAllLocal(root)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(issues) != 0 {
			t.Errorf("got %d issues, want 0", len(issues))
		}
	})

	t.Run("loads open issues", func(t *testing.T) {
		root := makeIssuesRoot(t, []issueFixture{
			{"1-open-issue.md", issue.Issue{Number: 1, Title: "Open issue", State: "open"}},
		})
		issues, err := loadAllLocal(root)
		if err != nil {
			t.Fatal(err)
		}
		if len(issues) != 1 {
			t.Fatalf("got %d issues, want 1", len(issues))
		}
		if issues[0].Title != "Open issue" {
			t.Errorf("Title: got %q, want %q", issues[0].Title, "Open issue")
		}
	})

	t.Run("loads closed issues", func(t *testing.T) {
		root := makeIssuesRoot(t, []issueFixture{
			{"1-closed-issue.md", issue.Issue{Number: 1, Title: "Closed issue", State: "closed"}},
		})
		issues, err := loadAllLocal(root)
		if err != nil {
			t.Fatal(err)
		}
		if len(issues) != 1 {
			t.Fatalf("got %d issues, want 1", len(issues))
		}
		if issues[0].State != "closed" {
			t.Errorf("State: got %q, want %q", issues[0].State, "closed")
		}
	})

	t.Run("loads both open and closed", func(t *testing.T) {
		root := makeIssuesRoot(t, []issueFixture{
			{"1-open.md", issue.Issue{Number: 1, Title: "Open", State: "open"}},
			{"2-closed.md", issue.Issue{Number: 2, Title: "Closed", State: "closed"}},
		})
		issues, err := loadAllLocal(root)
		if err != nil {
			t.Fatal(err)
		}
		if len(issues) != 2 {
			t.Errorf("got %d issues, want 2", len(issues))
		}
	})

	t.Run("skips non-md files", func(t *testing.T) {
		root := makeIssuesRoot(t, []issueFixture{
			{"1-issue.md", issue.Issue{Number: 1, Title: "Issue", State: "open"}},
		})
		if err := os.WriteFile(filepath.Join(root, "open", "readme.txt"), []byte("ignored"), 0644); err != nil {
			t.Fatal(err)
		}
		issues, err := loadAllLocal(root)
		if err != nil {
			t.Fatal(err)
		}
		if len(issues) != 1 {
			t.Errorf("got %d issues, want 1", len(issues))
		}
	})

	t.Run("skips colocated comments JSON files", func(t *testing.T) {
		root := makeIssuesRoot(t, []issueFixture{
			{"1-issue.md", issue.Issue{Number: 1, Title: "Issue", State: "open"}},
		})
		commentsJSON := filepath.Join(root, "open", "1-issue.comments.json")
		if err := os.WriteFile(commentsJSON, []byte(`[{"body":"a comment"}]`), 0644); err != nil {
			t.Fatal(err)
		}
		issues, err := loadAllLocal(root)
		if err != nil {
			t.Fatal(err)
		}
		if len(issues) != 1 {
			t.Errorf("got %d issues, want 1 (comments JSON should be ignored)", len(issues))
		}
	})

	t.Run("skips directories inside open", func(t *testing.T) {
		root := makeIssuesRoot(t, nil)
		if err := os.MkdirAll(filepath.Join(root, "open", "subdir.md"), 0755); err != nil {
			t.Fatal(err)
		}
		issues, err := loadAllLocal(root)
		if err != nil {
			t.Fatal(err)
		}
		if len(issues) != 0 {
			t.Errorf("got %d issues, want 0", len(issues))
		}
	})

	t.Run("local T-issues have Number zero", func(t *testing.T) {
		root := makeIssuesRoot(t, []issueFixture{
			{"T1-local-draft.md", issue.Issue{Title: "Local draft", State: "open"}},
		})
		issues, err := loadAllLocal(root)
		if err != nil {
			t.Fatal(err)
		}
		if len(issues) != 1 {
			t.Fatalf("got %d issues, want 1", len(issues))
		}
		if issues[0].Number != 0 {
			t.Errorf("Number: got %d, want 0", issues[0].Number)
		}
	})

	t.Run("sets Path on each issue", func(t *testing.T) {
		root := makeIssuesRoot(t, []issueFixture{
			{"1-issue.md", issue.Issue{Number: 1, Title: "Issue", State: "open"}},
		})
		issues, err := loadAllLocal(root)
		if err != nil {
			t.Fatal(err)
		}
		if len(issues) != 1 {
			t.Fatalf("got %d issues, want 1", len(issues))
		}
		if issues[0].Path == "" {
			t.Error("Path should not be empty")
		}
	})
}

func TestFindLocalByID(t *testing.T) {
	root := makeIssuesRoot(t, []issueFixture{
		{"1-first-issue.md", issue.Issue{Number: 1, Title: "First issue", State: "open"}},
		{"2-second-issue.md", issue.Issue{Number: 2, Title: "Second issue", State: "open"}},
		{"10-tenth-issue.md", issue.Issue{Number: 10, Title: "Tenth issue", State: "open"}},
		{"3-closed-issue.md", issue.Issue{Number: 3, Title: "Closed issue", State: "closed"}},
		{"T1-local-draft.md", issue.Issue{Title: "Local draft", State: "open"}},
		{"T2-another-draft.md", issue.Issue{Title: "Another draft", State: "open"}},
	})

	tests := []struct {
		id        string
		wantTitle string
		wantErr   bool
	}{
		{"1", "First issue", false},
		{"2", "Second issue", false},
		{"10", "Tenth issue", false},
		{"3", "Closed issue", false}, // finds closed issues
		{"T1", "Local draft", false},
		{"T2", "Another draft", false},
		{"t1", "Local draft", false}, // case-insensitive T-prefix
		{"t2", "Another draft", false},
		{"99", "", true},  // number not found
		{"T99", "", true}, // T-id not found
		{"abc", "", true}, // not numeric, not T-prefixed
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			iss, err := findLocalByID(root, tt.id)
			if tt.wantErr {
				if err == nil {
					t.Errorf("findLocalByID(%q): expected error, got nil", tt.id)
				}
				return
			}
			if err != nil {
				t.Fatalf("findLocalByID(%q): unexpected error: %v", tt.id, err)
			}
			if iss.Title != tt.wantTitle {
				t.Errorf("findLocalByID(%q): Title = %q, want %q", tt.id, iss.Title, tt.wantTitle)
			}
		})
	}
}
