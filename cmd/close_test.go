package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jamesjohnsdev/issues/internal/issue"
)

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
