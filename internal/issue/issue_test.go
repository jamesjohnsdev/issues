package issue

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeTemp writes content to a temp file and returns its path.
func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "issue-*.md")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return f.Name()
}

// checkIssue compares issue fields, excluding Path which is set from the file path.
func checkIssue(t *testing.T, got, want *Issue) {
	t.Helper()
	if got.Number != want.Number {
		t.Errorf("Number: got %d, want %d", got.Number, want.Number)
	}
	if got.Title != want.Title {
		t.Errorf("Title: got %q, want %q", got.Title, want.Title)
	}
	if got.State != want.State {
		t.Errorf("State: got %q, want %q", got.State, want.State)
	}
	if got.Body != want.Body {
		t.Errorf("Body: got %q, want %q", got.Body, want.Body)
	}
	if got.Milestone != want.Milestone {
		t.Errorf("Milestone: got %q, want %q", got.Milestone, want.Milestone)
	}
	if len(got.Labels) != len(want.Labels) {
		t.Errorf("Labels: got %v, want %v", got.Labels, want.Labels)
	} else {
		for i := range want.Labels {
			if got.Labels[i] != want.Labels[i] {
				t.Errorf("Labels[%d]: got %q, want %q", i, got.Labels[i], want.Labels[i])
			}
		}
	}
	if len(got.Assignees) != len(want.Assignees) {
		t.Errorf("Assignees: got %v, want %v", got.Assignees, want.Assignees)
	} else {
		for i := range want.Assignees {
			if got.Assignees[i] != want.Assignees[i] {
				t.Errorf("Assignees[%d]: got %q, want %q", i, got.Assignees[i], want.Assignees[i])
			}
		}
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    Issue
		wantErr bool
	}{
		{
			name:    "minimal valid issue",
			content: "---\ntitle: Hello\nstate: open\n---\n",
			want:    Issue{Title: "Hello", State: "open"},
		},
		{
			name:    "with body",
			content: "---\ntitle: Hello\nstate: open\n---\n\nsome body text\n",
			want:    Issue{Title: "Hello", State: "open", Body: "some body text\n"},
		},
		{
			name:    "with github number",
			content: "---\nnumber: 42\ntitle: My issue\nstate: open\n---\n",
			want:    Issue{Number: 42, Title: "My issue", State: "open"},
		},
		{
			name:    "closed state",
			content: "---\ntitle: Done\nstate: closed\n---\n",
			want:    Issue{Title: "Done", State: "closed"},
		},
		{
			name:    "with labels",
			content: "---\ntitle: Bug\nstate: open\nlabels:\n  - bug\n  - urgent\n---\n",
			want:    Issue{Title: "Bug", State: "open", Labels: []string{"bug", "urgent"}},
		},
		{
			name:    "with assignees",
			content: "---\ntitle: Task\nstate: open\nassignees:\n  - alice\n  - bob\n---\n",
			want:    Issue{Title: "Task", State: "open", Assignees: []string{"alice", "bob"}},
		},
		{
			name:    "with milestone",
			content: "---\ntitle: Feature\nstate: open\nmilestone: v1.0\n---\n",
			want:    Issue{Title: "Feature", State: "open", Milestone: "v1.0"},
		},
		{
			name:    "multiline body",
			content: "---\ntitle: Big issue\nstate: open\n---\n\nLine one.\n\nLine two.\n",
			want:    Issue{Title: "Big issue", State: "open", Body: "Line one.\n\nLine two.\n"},
		},
		{
			name:    "missing frontmatter delimiter",
			content: "no frontmatter here",
			wantErr: true,
		},
		{
			name:    "unclosed frontmatter",
			content: "---\ntitle: Hello\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeTemp(t, tt.content)
			got, err := Parse(path)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Path != path {
				t.Errorf("Path: got %q, want %q", got.Path, path)
			}
			checkIssue(t, got, &tt.want)
		})
	}

	t.Run("non-existent file returns error", func(t *testing.T) {
		_, err := Parse("/no/such/file.md")
		if err == nil {
			t.Error("expected error for non-existent file, got nil")
		}
	})
}

func TestWriteParse(t *testing.T) {
	tests := []struct {
		name string
		iss  Issue
	}{
		{
			name: "basic open issue",
			iss:  Issue{Title: "Hello", State: "open"},
		},
		{
			name: "issue with body",
			iss:  Issue{Title: "With body", State: "open", Body: "some body text\n"},
		},
		{
			name: "github issue with number",
			iss:  Issue{Number: 42, Title: "Numbered", State: "open"},
		},
		{
			name: "closed issue",
			iss:  Issue{Title: "Done", State: "closed"},
		},
		{
			name: "issue with labels",
			iss:  Issue{Title: "Labelled", State: "open", Labels: []string{"bug", "urgent"}},
		},
		{
			name: "issue with assignees",
			iss:  Issue{Title: "Assigned", State: "open", Assignees: []string{"alice", "bob"}},
		},
		{
			name: "issue with milestone",
			iss:  Issue{Title: "Milestoned", State: "open", Milestone: "v1.0"},
		},
		{
			name: "all fields",
			iss: Issue{
				Number:    7,
				Title:     "Full issue",
				State:     "open",
				Labels:    []string{"bug"},
				Assignees: []string{"alice"},
				Milestone: "v2.0",
				Body:      "body here\n",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "issue.md")
			if err := Write(path, &tt.iss); err != nil {
				t.Fatalf("Write: %v", err)
			}
			got, err := Parse(path)
			if err != nil {
				t.Fatalf("Parse after Write: %v", err)
			}
			checkIssue(t, got, &tt.iss)
		})
	}
}

func TestWriteTemplate(t *testing.T) {
	t.Run("empty issue shows all fields and separator", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "issue.md")
		iss := &Issue{State: "open"}
		if err := WriteTemplate(path, iss); err != nil {
			t.Fatalf("WriteTemplate: %v", err)
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		content := string(raw)
		for _, want := range []string{"title:", "state:", "labels:", "assignees:", "milestone:", "# Body goes below this line"} {
			if !strings.Contains(content, want) {
				t.Errorf("template missing %q\ngot:\n%s", want, content)
			}
		}
	})

	t.Run("parses back with empty body", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "issue.md")
		iss := &Issue{Title: "Draft", State: "open"}
		if err := WriteTemplate(path, iss); err != nil {
			t.Fatalf("WriteTemplate: %v", err)
		}
		got, err := Parse(path)
		if err != nil {
			t.Fatalf("Parse: %v", err)
		}
		if got.Title != "Draft" {
			t.Errorf("Title = %q, want %q", got.Title, "Draft")
		}
		if got.Body != "" {
			t.Errorf("Body = %q, want empty", got.Body)
		}
	})

	t.Run("preserves populated fields", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "issue.md")
		iss := &Issue{
			Title:     "My Issue",
			State:     "open",
			Labels:    []string{"bug", "urgent"},
			Assignees: []string{"alice"},
			Milestone: "v1.0",
		}
		if err := WriteTemplate(path, iss); err != nil {
			t.Fatalf("WriteTemplate: %v", err)
		}
		got, err := Parse(path)
		if err != nil {
			t.Fatalf("Parse: %v", err)
		}
		checkIssue(t, got, iss)
	})
}

func TestFilename(t *testing.T) {
	tests := []struct {
		name string
		iss  Issue
		want string
	}{
		{
			name: "basic numbered issue",
			iss:  Issue{Number: 1, Title: "Hello World"},
			want: "1-hello-world.md",
		},
		{
			name: "special chars in title",
			iss:  Issue{Number: 42, Title: "Fix: bug #42!"},
			want: "42-fix-bug-42.md",
		},
		{
			name: "double digit number",
			iss:  Issue{Number: 10, Title: "Ten"},
			want: "10-ten.md",
		},
		{
			name: "long title is slugged and truncated",
			iss:  Issue{Number: 3, Title: "this is a very long title that goes way beyond fifty characters"},
			want: "3-this-is-a-very-long-title-that-goes-way-beyond.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Filename(&tt.iss)
			if got != tt.want {
				t.Errorf("Filename() = %q, want %q", got, tt.want)
			}
		})
	}
}
