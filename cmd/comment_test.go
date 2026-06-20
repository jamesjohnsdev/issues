package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jamesjohnsdev/issues/internal/issue"
)

func TestComment(t *testing.T) {
	t.Run("no editor returns error", func(t *testing.T) {
		parent := makeProjectDir(t, []issueFixture{
			{"1-my-issue.md", issue.Issue{Number: 1, Title: "My issue", State: "open"}},
		})
		chdirTo(t, parent)
		t.Setenv("VISUAL", "")
		t.Setenv("EDITOR", "")

		err := commentCmd.RunE(commentCmd, []string{"1"})
		if err == nil {
			t.Error("expected error when no editor is set, got nil")
		}
	})

	t.Run("unknown issue returns error", func(t *testing.T) {
		parent := makeProjectDir(t, nil)
		chdirTo(t, parent)
		t.Setenv("VISUAL", "true")

		err := commentCmd.RunE(commentCmd, []string{"99"})
		if err == nil {
			t.Error("expected error for unknown issue, got nil")
		}
	})

	t.Run("editor that writes empty body discards comment", func(t *testing.T) {
		parent := makeProjectDir(t, []issueFixture{
			{"1-my-issue.md", issue.Issue{Number: 1, Title: "My issue", State: "open"}},
		})
		chdirTo(t, parent)
		t.Setenv("VISUAL", "")
		t.Setenv("EDITOR", "true") // no-op editor leaves file empty

		_ = captureStdout(t, func() {
			if err := commentCmd.RunE(commentCmd, []string{"1"}); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})

		path := filepath.Join(parent, issuesDirName, "open", "1-my-issue.comments.json")
		if _, err := os.Stat(path); err == nil {
			t.Error("expected no comments file after empty-body abort, but file exists")
		}
	})

	t.Run("editor that writes body saves draft to JSON", func(t *testing.T) {
		parent := makeProjectDir(t, []issueFixture{
			{"1-my-issue.md", issue.Issue{Number: 1, Title: "My issue", State: "open"}},
		})
		chdirTo(t, parent)

		script := filepath.Join(t.TempDir(), "editor.sh")
		if err := os.WriteFile(script, []byte("#!/bin/sh\necho 'hello world' > \"$1\"\n"), 0755); err != nil {
			t.Fatal(err)
		}
		t.Setenv("VISUAL", "")
		t.Setenv("EDITOR", script)

		_ = captureStdout(t, func() {
			if err := commentCmd.RunE(commentCmd, []string{"1"}); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})

		commentsFile := filepath.Join(parent, issuesDirName, "open", "1-my-issue.comments.json")
		comments, err := issue.ParseComments(commentsFile)
		if err != nil {
			t.Fatalf("ParseComments: %v", err)
		}
		if len(comments) != 1 {
			t.Fatalf("got %d comments, want 1", len(comments))
		}
		if comments[0].ID != "" {
			t.Errorf("expected empty ID for draft, got %q", comments[0].ID)
		}
		if comments[0].Body != "hello world" {
			t.Errorf("Body = %q, want %q", comments[0].Body, "hello world")
		}
	})

	t.Run("draft appended after existing comments", func(t *testing.T) {
		parent := makeProjectDir(t, []issueFixture{
			{"1-my-issue.md", issue.Issue{Number: 1, Title: "My issue", State: "open"}},
		})
		chdirTo(t, parent)

		// Pre-populate comments file with a synced comment
		ts := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		existing := []*issue.Comment{
			{ID: "IC_abc", Author: "alice", CreatedAt: &ts, Body: "existing comment"},
		}
		commentsFile := filepath.Join(parent, issuesDirName, "open", "1-my-issue.comments.json")
		if err := issue.WriteComments(commentsFile, existing); err != nil {
			t.Fatalf("WriteComments: %v", err)
		}

		script := filepath.Join(t.TempDir(), "editor.sh")
		if err := os.WriteFile(script, []byte("#!/bin/sh\necho 'new draft' > \"$1\"\n"), 0755); err != nil {
			t.Fatal(err)
		}
		t.Setenv("VISUAL", "")
		t.Setenv("EDITOR", script)

		_ = captureStdout(t, func() {
			if err := commentCmd.RunE(commentCmd, []string{"1"}); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})

		comments, err := issue.ParseComments(commentsFile)
		if err != nil {
			t.Fatalf("ParseComments: %v", err)
		}
		if len(comments) != 2 {
			t.Fatalf("got %d comments, want 2", len(comments))
		}
		if comments[0].ID != "IC_abc" {
			t.Errorf("existing comment should be first, got ID %q", comments[0].ID)
		}
		if comments[1].Body != "new draft" {
			t.Errorf("new draft should be last, got body %q", comments[1].Body)
		}
	})

	t.Run("whitespace-only body discards comment", func(t *testing.T) {
		parent := makeProjectDir(t, []issueFixture{
			{"1-my-issue.md", issue.Issue{Number: 1, Title: "My issue", State: "open"}},
		})
		chdirTo(t, parent)

		script := filepath.Join(t.TempDir(), "editor.sh")
		if err := os.WriteFile(script, []byte("#!/bin/sh\nprintf '   \\n' > \"$1\"\n"), 0755); err != nil {
			t.Fatal(err)
		}
		t.Setenv("VISUAL", "")
		t.Setenv("EDITOR", script)

		_ = captureStdout(t, func() {
			if err := commentCmd.RunE(commentCmd, []string{"1"}); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})

		path := filepath.Join(parent, issuesDirName, "open", "1-my-issue.comments.json")
		if _, err := os.Stat(path); err == nil {
			t.Error("expected no comments file after whitespace-only body, but file exists")
		}
	})
}
