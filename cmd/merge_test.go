package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesjohnsdev/issues/internal/issue"
)

func TestMergeBothLocal(t *testing.T) {
	t.Run("closes a in closed dir with state closed", func(t *testing.T) {
		parent := makeProjectDir(t, []issueFixture{
			{"T1-issue-a.md", issue.Issue{Title: "Issue A", State: "open"}},
			{"T2-issue-b.md", issue.Issue{Title: "Issue B", State: "open"}},
		})
		chdirTo(t, parent)
		issuesDir := filepath.Join(parent, issuesDirName)

		_ = captureStdout(t, func() {
			if err := mergeCmd.RunE(mergeCmd, []string{"T1", "T2"}); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})

		closedPath := filepath.Join(issuesDir, "closed", "T1-issue-a.md")
		iss, err := issue.Parse(closedPath)
		if err != nil {
			t.Fatalf("parsing closed issue: %v", err)
		}
		if iss.State != "closed" {
			t.Errorf("State = %q, want closed", iss.State)
		}
	})

	t.Run("removes a from open dir", func(t *testing.T) {
		parent := makeProjectDir(t, []issueFixture{
			{"T1-issue-a.md", issue.Issue{Title: "Issue A", State: "open"}},
			{"T2-issue-b.md", issue.Issue{Title: "Issue B", State: "open"}},
		})
		chdirTo(t, parent)
		issuesDir := filepath.Join(parent, issuesDirName)

		_ = captureStdout(t, func() {
			if err := mergeCmd.RunE(mergeCmd, []string{"T1", "T2"}); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})

		if _, err := os.Stat(filepath.Join(issuesDir, "open", "T1-issue-a.md")); !os.IsNotExist(err) {
			t.Error("expected T1-issue-a.md to be removed from open/")
		}
	})

	t.Run("leaves b unchanged in open dir", func(t *testing.T) {
		parent := makeProjectDir(t, []issueFixture{
			{"T1-issue-a.md", issue.Issue{Title: "Issue A", State: "open"}},
			{"T2-issue-b.md", issue.Issue{Title: "Issue B", State: "open"}},
		})
		chdirTo(t, parent)
		issuesDir := filepath.Join(parent, issuesDirName)

		_ = captureStdout(t, func() {
			if err := mergeCmd.RunE(mergeCmd, []string{"T1", "T2"}); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})

		if _, err := os.Stat(filepath.Join(issuesDir, "open", "T2-issue-b.md")); os.IsNotExist(err) {
			t.Error("expected T2-issue-b.md to remain in open/")
		}
	})

	t.Run("writes duplicate comment referencing b on a's comments file", func(t *testing.T) {
		parent := makeProjectDir(t, []issueFixture{
			{"T1-issue-a.md", issue.Issue{Title: "Issue A", State: "open"}},
			{"T2-issue-b.md", issue.Issue{Title: "Issue B", State: "open"}},
		})
		chdirTo(t, parent)
		issuesDir := filepath.Join(parent, issuesDirName)

		_ = captureStdout(t, func() {
			if err := mergeCmd.RunE(mergeCmd, []string{"T1", "T2"}); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})

		commentsPath := filepath.Join(issuesDir, "closed", "0-issue-a.comments.json")
		comments, err := issue.ParseComments(commentsPath)
		if err != nil {
			t.Fatalf("ParseComments: %v", err)
		}
		if len(comments) != 1 {
			t.Fatalf("expected 1 comment, got %d", len(comments))
		}
		body := comments[0].Body
		if !strings.Contains(body, "T2") {
			t.Errorf("comment body %q should reference T2", body)
		}
		if !strings.Contains(body, "Issue B") {
			t.Errorf("comment body %q should reference b's title", body)
		}
	})

	t.Run("comment has nil metadata (local draft)", func(t *testing.T) {
		parent := makeProjectDir(t, []issueFixture{
			{"T1-issue-a.md", issue.Issue{Title: "Issue A", State: "open"}},
			{"T2-issue-b.md", issue.Issue{Title: "Issue B", State: "open"}},
		})
		chdirTo(t, parent)
		issuesDir := filepath.Join(parent, issuesDirName)

		_ = captureStdout(t, func() {
			if err := mergeCmd.RunE(mergeCmd, []string{"T1", "T2"}); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})

		commentsPath := filepath.Join(issuesDir, "closed", "0-issue-a.comments.json")
		comments, err := issue.ParseComments(commentsPath)
		if err != nil {
			t.Fatalf("ParseComments: %v", err)
		}
		if len(comments) == 0 {
			t.Fatal("no comments")
		}
		if comments[0].Metadata != nil {
			t.Errorf("expected nil Metadata for local draft comment, got %+v", comments[0].Metadata)
		}
	})

	t.Run("comments file moved to closed dir", func(t *testing.T) {
		parent := makeProjectDir(t, []issueFixture{
			{"T1-issue-a.md", issue.Issue{Title: "Issue A", State: "open"}},
			{"T2-issue-b.md", issue.Issue{Title: "Issue B", State: "open"}},
		})
		chdirTo(t, parent)
		issuesDir := filepath.Join(parent, issuesDirName)

		_ = captureStdout(t, func() {
			if err := mergeCmd.RunE(mergeCmd, []string{"T1", "T2"}); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})

		if _, err := os.Stat(filepath.Join(issuesDir, "open", "0-issue-a.comments.json")); !os.IsNotExist(err) {
			t.Error("expected comments file to be removed from open/")
		}
		if _, err := os.Stat(filepath.Join(issuesDir, "closed", "0-issue-a.comments.json")); os.IsNotExist(err) {
			t.Error("expected comments file to exist in closed/")
		}
	})

	t.Run("preserves pre-existing comments on a", func(t *testing.T) {
		parent := makeProjectDir(t, []issueFixture{
			{"T1-issue-a.md", issue.Issue{Title: "Issue A", State: "open"}},
			{"T2-issue-b.md", issue.Issue{Title: "Issue B", State: "open"}},
		})
		chdirTo(t, parent)
		issuesDir := filepath.Join(parent, issuesDirName)

		existing := []*issue.Comment{{Body: "pre-existing comment"}}
		commentsFile := filepath.Join(issuesDir, "open", "0-issue-a.comments.json")
		if err := issue.WriteComments(commentsFile, existing); err != nil {
			t.Fatalf("WriteComments: %v", err)
		}

		_ = captureStdout(t, func() {
			if err := mergeCmd.RunE(mergeCmd, []string{"T1", "T2"}); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})

		closedComments := filepath.Join(issuesDir, "closed", "0-issue-a.comments.json")
		comments, err := issue.ParseComments(closedComments)
		if err != nil {
			t.Fatalf("ParseComments: %v", err)
		}
		if len(comments) != 2 {
			t.Fatalf("expected 2 comments (existing + merge note), got %d", len(comments))
		}
		if comments[0].Body != "pre-existing comment" {
			t.Errorf("first comment = %q, want pre-existing comment", comments[0].Body)
		}
	})

	t.Run("case-insensitive T-ids", func(t *testing.T) {
		parent := makeProjectDir(t, []issueFixture{
			{"T1-issue-a.md", issue.Issue{Title: "Issue A", State: "open"}},
			{"T2-issue-b.md", issue.Issue{Title: "Issue B", State: "open"}},
		})
		chdirTo(t, parent)
		issuesDir := filepath.Join(parent, issuesDirName)

		_ = captureStdout(t, func() {
			if err := mergeCmd.RunE(mergeCmd, []string{"t1", "t2"}); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})

		if _, err := os.Stat(filepath.Join(issuesDir, "closed", "T1-issue-a.md")); os.IsNotExist(err) {
			t.Error("expected T1-issue-a.md in closed/ after lowercase t-id merge")
		}
	})
}

func TestMergeErrors(t *testing.T) {
	t.Run("unknown issue a returns error", func(t *testing.T) {
		parent := makeProjectDir(t, []issueFixture{
			{"T2-issue-b.md", issue.Issue{Title: "Issue B", State: "open"}},
		})
		chdirTo(t, parent)

		err := mergeCmd.RunE(mergeCmd, []string{"T1", "T2"})
		if err == nil {
			t.Error("expected error for unknown issue a, got nil")
		}
	})

	t.Run("unknown issue b returns error", func(t *testing.T) {
		parent := makeProjectDir(t, []issueFixture{
			{"T1-issue-a.md", issue.Issue{Title: "Issue A", State: "open"}},
		})
		chdirTo(t, parent)

		err := mergeCmd.RunE(mergeCmd, []string{"T1", "T2"})
		if err == nil {
			t.Error("expected error for unknown issue b, got nil")
		}
	})

	t.Run("invalid issue id returns error", func(t *testing.T) {
		parent := makeProjectDir(t, nil)
		chdirTo(t, parent)

		err := mergeCmd.RunE(mergeCmd, []string{"abc", "T2"})
		if err == nil {
			t.Error("expected error for invalid id, got nil")
		}
	})
}

func TestMergeOnlineLocalAbort(t *testing.T) {
	t.Run("answering N leaves both issues unchanged", func(t *testing.T) {
		parent := makeProjectDir(t, []issueFixture{
			{"5-github-issue.md", issue.Issue{Number: 5, Title: "GitHub issue", State: "open"}},
			{"T1-local-issue.md", issue.Issue{Title: "Local issue", State: "open"}},
		})
		chdirTo(t, parent)
		issuesDir := filepath.Join(parent, issuesDirName)
		injectStdin(t, "n\n")

		_ = captureStdout(t, func() {
			if err := mergeCmd.RunE(mergeCmd, []string{"5", "T1"}); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})

		if _, err := os.Stat(filepath.Join(issuesDir, "open", "5-github-issue.md")); os.IsNotExist(err) {
			t.Error("expected GitHub issue to remain in open/ after abort")
		}
		commentsPath := filepath.Join(issuesDir, "open", "0-local-issue.comments.json")
		if _, err := os.Stat(commentsPath); err == nil {
			t.Error("expected no comments file to be created after abort")
		}
	})

	t.Run("empty answer is treated as N", func(t *testing.T) {
		parent := makeProjectDir(t, []issueFixture{
			{"5-github-issue.md", issue.Issue{Number: 5, Title: "GitHub issue", State: "open"}},
			{"T1-local-issue.md", issue.Issue{Title: "Local issue", State: "open"}},
		})
		chdirTo(t, parent)
		issuesDir := filepath.Join(parent, issuesDirName)
		injectStdin(t, "\n")

		_ = captureStdout(t, func() {
			if err := mergeCmd.RunE(mergeCmd, []string{"5", "T1"}); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})

		if _, err := os.Stat(filepath.Join(issuesDir, "open", "5-github-issue.md")); os.IsNotExist(err) {
			t.Error("expected GitHub issue to remain in open/ after empty-answer abort")
		}
	})
}
