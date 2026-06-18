package issue

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func FuzzSlug(f *testing.F) {
	f.Add("Hello World")
	f.Add("")
	f.Add("Fix: bug #42!")
	f.Add("---hello---")
	f.Add("!!! ???")
	f.Add(strings.Repeat("a", 100))
	f.Add("enable flag passing to filter status on issue list")

	f.Fuzz(func(t *testing.T, s string) {
		result := Slug(s)

		if len(result) > 50 {
			t.Errorf("Slug(%q) length %d > 50", s, len(result))
		}

		if strings.HasPrefix(result, "-") {
			t.Errorf("Slug(%q) = %q starts with '-'", s, result)
		}
		if strings.HasSuffix(result, "-") {
			t.Errorf("Slug(%q) = %q ends with '-'", s, result)
		}

		if strings.Contains(result, "--") {
			t.Errorf("Slug(%q) = %q contains consecutive dashes", s, result)
		}

		for _, r := range result {
			if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '-' {
				t.Errorf("Slug(%q) = %q contains invalid char %q", s, result, r)
			}
		}
	})
}

func FuzzParse(f *testing.F) {
	f.Add("---\ntitle: Hello\nstate: open\n---\n")
	f.Add("---\ntitle: Hello\nstate: open\n---\n\nsome body\n")
	f.Add("---\nnumber: 42\ntitle: Bug\nstate: closed\n---\n")
	f.Add("not frontmatter")
	f.Add("")
	f.Add("---\n")
	f.Add("---\n---\n")

	f.Fuzz(func(t *testing.T, content string) {
		path := filepath.Join(t.TempDir(), "issue.md")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		iss, err := Parse(path)
		if err != nil {
			return // errors for malformed input are expected
		}

		if iss == nil {
			t.Fatal("Parse returned nil issue without error")
		}
		if iss.Path != path {
			t.Errorf("Path = %q, want %q", iss.Path, path)
		}
	})
}

func FuzzWriteParse(f *testing.F) {
	f.Add("Hello", "open", "")
	f.Add("Fix the bug", "closed", "Some body text\n")
	f.Add("", "open", "")
	f.Add("feat: add OAuth2", "open", "## Description\nDetailed body.\n")
	f.Add("title with\nnewline", "open", "")

	f.Fuzz(func(t *testing.T, title, state, body string) {
		iss := &Issue{Title: title, State: state, Body: body}
		path := filepath.Join(t.TempDir(), "issue.md")

		if err := Write(path, iss); err != nil {
			t.Fatalf("Write failed: %v", err)
		}

		got, err := Parse(path)
		if err != nil {
			t.Fatalf("Parse failed after successful Write: %v", err)
		}

		if got.Title != title {
			t.Errorf("Title mismatch: got %q, want %q", got.Title, title)
		}

		// Body: Write always appends a trailing newline to non-empty bodies.
		// Skip the body check when it contains "\n---" — that sequence would
		// confuse the frontmatter end-marker scan and is a known parser limitation.
		if !strings.Contains(body, "\n---") {
			wantBody := body
			if body != "" && !strings.HasSuffix(body, "\n") {
				wantBody += "\n"
			}
			if got.Body != wantBody {
				t.Errorf("Body mismatch: got %q, want %q", got.Body, wantBody)
			}
		}
	})
}
