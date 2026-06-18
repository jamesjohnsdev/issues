package cmd

import (
	"strings"
	"testing"

	"github.com/jamesjohnsdev/issues/internal/issue"
)

func TestList(t *testing.T) {
	t.Run("default shows only open issues", func(t *testing.T) {
		parent := makeProjectDir(t, []issueFixture{
			{"1-open.md", issue.Issue{Number: 1, Title: "Open issue", State: "open"}},
			{"2-closed.md", issue.Issue{Number: 2, Title: "Closed issue", State: "closed"}},
		})
		chdirTo(t, parent)
		t.Cleanup(func() { listAll = false; listClosed = false })

		out := captureStdout(t, func() {
			if err := listCmd.RunE(listCmd, nil); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})

		if !strings.Contains(out, "Open issue") {
			t.Error("expected output to contain open issue title")
		}
		if strings.Contains(out, "Closed issue") {
			t.Error("expected output to not contain closed issue title")
		}
	})

	t.Run("closed flag shows only closed issues", func(t *testing.T) {
		parent := makeProjectDir(t, []issueFixture{
			{"1-open.md", issue.Issue{Number: 1, Title: "Open issue", State: "open"}},
			{"2-closed.md", issue.Issue{Number: 2, Title: "Closed issue", State: "closed"}},
		})
		chdirTo(t, parent)
		listClosed = true
		t.Cleanup(func() { listAll = false; listClosed = false })

		out := captureStdout(t, func() {
			if err := listCmd.RunE(listCmd, nil); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})

		if strings.Contains(out, "Open issue") {
			t.Error("expected output to not contain open issue title")
		}
		if !strings.Contains(out, "Closed issue") {
			t.Error("expected output to contain closed issue title")
		}
	})

	t.Run("all flag shows both open and closed", func(t *testing.T) {
		parent := makeProjectDir(t, []issueFixture{
			{"1-open.md", issue.Issue{Number: 1, Title: "Open issue", State: "open"}},
			{"2-closed.md", issue.Issue{Number: 2, Title: "Closed issue", State: "closed"}},
		})
		chdirTo(t, parent)
		listAll = true
		t.Cleanup(func() { listAll = false; listClosed = false })

		out := captureStdout(t, func() {
			if err := listCmd.RunE(listCmd, nil); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})

		if !strings.Contains(out, "Open issue") {
			t.Error("expected output to contain open issue title")
		}
		if !strings.Contains(out, "Closed issue") {
			t.Error("expected output to contain closed issue title")
		}
	})

	t.Run("no matching issues prints no-issues message", func(t *testing.T) {
		parent := makeProjectDir(t, nil)
		chdirTo(t, parent)
		t.Cleanup(func() { listAll = false; listClosed = false })

		out := captureStdout(t, func() {
			if err := listCmd.RunE(listCmd, nil); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})

		if !strings.Contains(out, "No local issues") {
			t.Errorf("expected no-issues message, got: %q", out)
		}
	})

	t.Run("issues sorted numerically not lexicographically", func(t *testing.T) {
		parent := makeProjectDir(t, []issueFixture{
			{"1-first.md", issue.Issue{Number: 1, Title: "First", State: "open"}},
			{"2-second.md", issue.Issue{Number: 2, Title: "Second", State: "open"}},
			{"10-tenth.md", issue.Issue{Number: 10, Title: "Tenth", State: "open"}},
		})
		chdirTo(t, parent)
		t.Cleanup(func() { listAll = false; listClosed = false })

		out := captureStdout(t, func() {
			if err := listCmd.RunE(listCmd, nil); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})

		posFirst := strings.Index(out, "First")
		posSecond := strings.Index(out, "Second")
		posTenth := strings.Index(out, "Tenth")

		if posFirst < 0 || posSecond < 0 || posTenth < 0 {
			t.Fatalf("not all issues in output: %q", out)
		}
		if posFirst >= posSecond || posSecond >= posTenth {
			t.Errorf("wrong order: First@%d Second@%d Tenth@%d — want First < Second < Tenth",
				posFirst, posSecond, posTenth)
		}
	})

	t.Run("T-issues appear after GitHub issues", func(t *testing.T) {
		parent := makeProjectDir(t, []issueFixture{
			{"1-github.md", issue.Issue{Number: 1, Title: "GitHub issue", State: "open"}},
			{"T1-local.md", issue.Issue{Title: "Local draft", State: "open"}},
		})
		chdirTo(t, parent)
		t.Cleanup(func() { listAll = false; listClosed = false })

		out := captureStdout(t, func() {
			if err := listCmd.RunE(listCmd, nil); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})

		posGH := strings.Index(out, "GitHub issue")
		posLocal := strings.Index(out, "Local draft")

		if posGH < 0 || posLocal < 0 {
			t.Fatalf("not all issues in output: %q", out)
		}
		if posGH > posLocal {
			t.Error("expected GitHub issue to appear before local T-issue")
		}
	})
}
