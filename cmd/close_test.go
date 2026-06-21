package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jamesjohnsdev/issues/internal/issue"
)

func setCloseComment(t *testing.T, val bool) {
	t.Helper()
	v := "false"
	if val {
		v = "true"
	}
	if err := closeCmd.Flags().Set("comment", v); err != nil {
		t.Fatalf("setting --comment flag: %v", err)
	}
	t.Cleanup(func() { _ = closeCmd.Flags().Set("comment", "false") })
}

func TestClose(t *testing.T) {
	tests := []struct {
		name             string
		arg              string
		fixtures         []issueFixture
		wantErr          bool
		wantInClosed     string // filename that should exist in closed/ after the command
		wantGoneFromOpen string // filename that should be gone from open/ after the command
	}{
		{
			name: "closes open GitHub issue by number",
			arg:  "1",
			fixtures: []issueFixture{
				{"1-first-issue.md", issue.Issue{Number: 1, Title: "First issue", State: "open"}},
			},
			wantInClosed:     "1-first-issue.md",
			wantGoneFromOpen: "1-first-issue.md",
		},
		{
			name: "closes open T-issue by T-id",
			arg:  "T1",
			fixtures: []issueFixture{
				{"T1-local-draft.md", issue.Issue{Title: "Local draft", State: "open"}},
			},
			wantInClosed:     "T1-local-draft.md",
			wantGoneFromOpen: "T1-local-draft.md",
		},
		{
			name: "case-insensitive T-id",
			arg:  "t1",
			fixtures: []issueFixture{
				{"T1-local-draft.md", issue.Issue{Title: "Local draft", State: "open"}},
			},
			wantInClosed:     "T1-local-draft.md",
			wantGoneFromOpen: "T1-local-draft.md",
		},
		{
			name: "already closed issue is a no-op",
			arg:  "1",
			fixtures: []issueFixture{
				{"1-done.md", issue.Issue{Number: 1, Title: "Done", State: "closed"}},
			},
		},
		{
			name:     "unknown number returns error",
			arg:      "99",
			fixtures: nil,
			wantErr:  true,
		},
		{
			name:     "unknown T-id returns error",
			arg:      "T99",
			fixtures: nil,
			wantErr:  true,
		},
		{
			name:     "invalid id returns error",
			arg:      "abc",
			fixtures: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parent := makeProjectDir(t, tt.fixtures)
			chdirTo(t, parent)
			issuesDir := filepath.Join(parent, issuesDirName)

			err := closeCmd.RunE(closeCmd, []string{tt.arg})
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantInClosed != "" {
				closedPath := filepath.Join(issuesDir, "closed", tt.wantInClosed)
				if _, err := os.Stat(closedPath); os.IsNotExist(err) {
					t.Errorf("expected %s in closed/, but it does not exist", tt.wantInClosed)
				} else {
					iss, err := issue.Parse(closedPath)
					if err != nil {
						t.Fatalf("parsing closed file: %v", err)
					}
					if iss.State != "closed" {
						t.Errorf("State = %q, want %q", iss.State, "closed")
					}
				}
			}

			if tt.wantGoneFromOpen != "" {
				openPath := filepath.Join(issuesDir, "open", tt.wantGoneFromOpen)
				if _, err := os.Stat(openPath); !os.IsNotExist(err) {
					t.Errorf("expected %s to be gone from open/, but it still exists", tt.wantGoneFromOpen)
				}
			}
		})
	}
}

func TestCloseMovesCommentsFile(t *testing.T) {
	parent := makeProjectDir(t, []issueFixture{
		{"1-my-issue.md", issue.Issue{Number: 1, Title: "My issue", State: "open"}},
	})
	chdirTo(t, parent)

	commentsFile := filepath.Join(parent, issuesDirName, "open", "1-my-issue.comments.json")
	if err := issue.WriteComments(commentsFile, []*issue.Comment{{Body: "existing"}}); err != nil {
		t.Fatal(err)
	}

	_ = captureStdout(t, func() {
		if err := closeCmd.RunE(closeCmd, []string{"1"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	closedComments := filepath.Join(parent, issuesDirName, "closed", "1-my-issue.comments.json")
	if _, err := os.Stat(closedComments); os.IsNotExist(err) {
		t.Error("expected comments file to be moved to closed/, but it was not found there")
	}
	if _, err := os.Stat(commentsFile); !os.IsNotExist(err) {
		t.Error("expected comments file to be gone from open/, but it still exists")
	}
}

func TestCloseWithComment(t *testing.T) {
	t.Run("no editor returns error", func(t *testing.T) {
		parent := makeProjectDir(t, []issueFixture{
			{"1-my-issue.md", issue.Issue{Number: 1, Title: "My issue", State: "open"}},
		})
		chdirTo(t, parent)
		t.Setenv("VISUAL", "")
		t.Setenv("EDITOR", "")
		setCloseComment(t, true)

		err := closeCmd.RunE(closeCmd, []string{"1"})
		if err == nil {
			t.Error("expected error when no editor set, got nil")
		}
	})

	t.Run("empty body aborts without closing", func(t *testing.T) {
		parent := makeProjectDir(t, []issueFixture{
			{"1-my-issue.md", issue.Issue{Number: 1, Title: "My issue", State: "open"}},
		})
		chdirTo(t, parent)
		t.Setenv("VISUAL", "")
		t.Setenv("EDITOR", "true")
		setCloseComment(t, true)

		_ = captureStdout(t, func() {
			if err := closeCmd.RunE(closeCmd, []string{"1"}); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})

		openPath := filepath.Join(parent, issuesDirName, "open", "1-my-issue.md")
		if _, err := os.Stat(openPath); os.IsNotExist(err) {
			t.Error("issue should not have been closed when comment body was empty")
		}
	})

	t.Run("comment saved and issue closed", func(t *testing.T) {
		parent := makeProjectDir(t, []issueFixture{
			{"1-my-issue.md", issue.Issue{Number: 1, Title: "My issue", State: "open"}},
		})
		chdirTo(t, parent)

		script := filepath.Join(t.TempDir(), "editor.sh")
		if err := os.WriteFile(script, []byte("#!/bin/sh\necho 'closing note' > \"$1\"\n"), 0755); err != nil {
			t.Fatal(err)
		}
		t.Setenv("VISUAL", "")
		t.Setenv("EDITOR", script)
		setCloseComment(t, true)

		_ = captureStdout(t, func() {
			if err := closeCmd.RunE(closeCmd, []string{"1"}); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})

		closedPath := filepath.Join(parent, issuesDirName, "closed", "1-my-issue.md")
		if _, err := os.Stat(closedPath); os.IsNotExist(err) {
			t.Error("expected issue in closed/ but it was not found")
		}

		commentsPath := filepath.Join(parent, issuesDirName, "closed", "1-my-issue.comments.json")
		comments, err := issue.ParseComments(commentsPath)
		if err != nil {
			t.Fatalf("ParseComments: %v", err)
		}
		if len(comments) != 1 {
			t.Fatalf("got %d comments, want 1", len(comments))
		}
		if comments[0].Body != "closing note" {
			t.Errorf("Body = %q, want %q", comments[0].Body, "closing note")
		}
	})
}
