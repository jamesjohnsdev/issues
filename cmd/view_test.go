package cmd

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jamesjohnsdev/issues/internal/issue"
)

func resetViewFlags(t *testing.T) {
	t.Helper()
	origComments := viewCommentsFlag
	origWeb := viewWebFlag
	t.Cleanup(func() {
		viewCommentsFlag = origComments
		viewWebFlag = origWeb
	})
	viewCommentsFlag = false
	viewWebFlag = false
}

func TestViewComments(t *testing.T) {
	t.Run("unknown issue returns error", func(t *testing.T) {
		parent := makeProjectDir(t, nil)
		chdirTo(t, parent)
		resetViewFlags(t)
		viewCommentsFlag = true

		err := viewCmd.RunE(viewCmd, []string{"99"})
		if err == nil {
			t.Error("expected error for unknown issue, got nil")
		}
	})

	t.Run("no comments file prints empty message", func(t *testing.T) {
		parent := makeProjectDir(t, []issueFixture{
			{"1-my-issue.md", issue.Issue{Number: 1, Title: "My issue", State: "open"}},
		})
		chdirTo(t, parent)
		resetViewFlags(t)
		viewCommentsFlag = true

		out := captureStdout(t, func() {
			if err := viewCmd.RunE(viewCmd, []string{"1"}); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})

		if !strings.Contains(out, "no comments") {
			t.Errorf("expected 'no comments' in output, got: %q", out)
		}
	})

	t.Run("synced comment shows author and body", func(t *testing.T) {
		parent := makeProjectDir(t, []issueFixture{
			{"1-my-issue.md", issue.Issue{Number: 1, Title: "My issue", State: "open"}},
		})
		chdirTo(t, parent)

		ts := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)
		comments := []*issue.Comment{
			{Metadata: &issue.CommentMeta{ID: "IC_abc", Author: "alice", CreatedAt: &ts}, Body: "looks good"},
		}
		commentsFile := filepath.Join(parent, issuesDirName, "open", "1-my-issue.comments.json")
		if err := issue.WriteComments(commentsFile, comments); err != nil {
			t.Fatal(err)
		}

		resetViewFlags(t)
		viewCommentsFlag = true

		out := captureStdout(t, func() {
			if err := viewCmd.RunE(viewCmd, []string{"1"}); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})

		if !strings.Contains(out, "alice") {
			t.Errorf("expected author 'alice' in output, got: %q", out)
		}
		if !strings.Contains(out, "looks good") {
			t.Errorf("expected body 'looks good' in output, got: %q", out)
		}
	})

	t.Run("local draft shows draft label", func(t *testing.T) {
		parent := makeProjectDir(t, []issueFixture{
			{"1-my-issue.md", issue.Issue{Number: 1, Title: "My issue", State: "open"}},
		})
		chdirTo(t, parent)

		comments := []*issue.Comment{
			{Body: "my pending comment"},
		}
		commentsFile := filepath.Join(parent, issuesDirName, "open", "1-my-issue.comments.json")
		if err := issue.WriteComments(commentsFile, comments); err != nil {
			t.Fatal(err)
		}

		resetViewFlags(t)
		viewCommentsFlag = true

		out := captureStdout(t, func() {
			if err := viewCmd.RunE(viewCmd, []string{"1"}); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})

		if !strings.Contains(out, "draft") {
			t.Errorf("expected 'draft' label in output for local comment, got: %q", out)
		}
		if !strings.Contains(out, "my pending comment") {
			t.Errorf("expected body in output, got: %q", out)
		}
	})

	t.Run("mixed comments preserves order", func(t *testing.T) {
		parent := makeProjectDir(t, []issueFixture{
			{"1-my-issue.md", issue.Issue{Number: 1, Title: "My issue", State: "open"}},
		})
		chdirTo(t, parent)

		ts := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
		comments := []*issue.Comment{
			{Metadata: &issue.CommentMeta{ID: "IC_1", Author: "bob", CreatedAt: &ts}, Body: "first"},
			{Body: "second draft"},
		}
		commentsFile := filepath.Join(parent, issuesDirName, "open", "1-my-issue.comments.json")
		if err := issue.WriteComments(commentsFile, comments); err != nil {
			t.Fatal(err)
		}

		resetViewFlags(t)
		viewCommentsFlag = true

		out := captureStdout(t, func() {
			if err := viewCmd.RunE(viewCmd, []string{"1"}); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})

		firstIdx := strings.Index(out, "first")
		secondIdx := strings.Index(out, "second draft")
		if firstIdx == -1 || secondIdx == -1 {
			t.Fatalf("expected both comment bodies in output, got: %q", out)
		}
		if firstIdx > secondIdx {
			t.Errorf("expected 'first' before 'second draft' in output")
		}
	})

	t.Run("without flag opens editor", func(t *testing.T) {
		parent := makeProjectDir(t, []issueFixture{
			{"1-my-issue.md", issue.Issue{Number: 1, Title: "My issue", State: "open"}},
		})
		chdirTo(t, parent)
		resetViewFlags(t) // viewCommentsFlag = false

		// Use a no-op editor so the test doesn't hang
		t.Setenv("VISUAL", "")
		t.Setenv("EDITOR", "true")

		if err := viewCmd.RunE(viewCmd, []string{"1"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestViewWeb(t *testing.T) {
	t.Run("web flag on local T-issue returns error", func(t *testing.T) {
		parent := makeProjectDir(t, []issueFixture{
			{"T1-local-only.md", issue.Issue{Number: 0, Title: "Local only", State: "open"}},
		})
		chdirTo(t, parent)
		resetViewFlags(t)
		viewWebFlag = true

		err := viewCmd.RunE(viewCmd, []string{"T1"})
		if err == nil {
			t.Error("expected error for local-only issue with --web, got nil")
		}
		if !strings.Contains(err.Error(), "local-only") {
			t.Errorf("expected 'local-only' in error, got: %v", err)
		}
	})
}
