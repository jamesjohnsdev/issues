package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jamesjohnsdev/issues/internal/issue"
)

func TestDeleteCmd(t *testing.T) {
	t.Run("local-only issue is deleted", func(t *testing.T) {
		parent := makeProjectDir(t, []issueFixture{
			{"T1-local-bug.md", issue.Issue{Title: "Local bug", State: "open"}},
		})
		chdirTo(t, parent)
		injectStdin(t, "")
		_ = captureStdout(t, func() {
			if err := deleteCmd.RunE(deleteCmd, []string{"T1"}); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		files := readMDFiles(t, filepath.Join(parent, issuesDirName, "open"))
		if len(files) != 0 {
			t.Errorf("expected file deleted, got %d files", len(files))
		}
	})

	t.Run("synced issue snapshot deleted after successful local delete", func(t *testing.T) {
		parent := makeProjectDir(t, []issueFixture{
			{"5-synced-bug.md", issue.Issue{Number: 5, Title: "Synced bug", State: "open"}},
		})
		issRoot := filepath.Join(parent, issuesDirName)
		origDir := originalsDir(issRoot)
		if err := os.MkdirAll(origDir, 0755); err != nil {
			t.Fatal(err)
		}
		origPath := filepath.Join(origDir, "5.md")
		if err := os.WriteFile(origPath, []byte("snapshot"), 0644); err != nil {
			t.Fatal(err)
		}
		chdirTo(t, parent)
		injectStdin(t, "n\n")
		_ = captureStdout(t, func() {
			if err := deleteCmd.RunE(deleteCmd, []string{"5"}); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if files := readMDFiles(t, filepath.Join(parent, issuesDirName, "open")); len(files) != 0 {
			t.Errorf("expected local file deleted, got %d files", len(files))
		}
		if _, err := os.Stat(origPath); !os.IsNotExist(err) {
			t.Error("expected snapshot deleted, but it still exists")
		}
	})

	t.Run("synced issue snapshot preserved when local delete fails", func(t *testing.T) {
		parent := makeProjectDir(t, []issueFixture{
			{"5-synced-bug.md", issue.Issue{Number: 5, Title: "Synced bug", State: "open"}},
		})
		issRoot := filepath.Join(parent, issuesDirName)
		origDir := originalsDir(issRoot)
		if err := os.MkdirAll(origDir, 0755); err != nil {
			t.Fatal(err)
		}
		origPath := filepath.Join(origDir, "5.md")
		if err := os.WriteFile(origPath, []byte("snapshot"), 0644); err != nil {
			t.Fatal(err)
		}
		openDir := filepath.Join(parent, issuesDirName, "open")
		if err := os.Chmod(openDir, 0555); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.Chmod(openDir, 0755) })
		chdirTo(t, parent)
		injectStdin(t, "n\n")
		var deleteErr error
		_ = captureStdout(t, func() {
			deleteErr = deleteCmd.RunE(deleteCmd, []string{"5"})
		})
		if deleteErr == nil {
			t.Fatal("expected error from failed local delete, got nil")
		}
		if _, err := os.Stat(origPath); os.IsNotExist(err) {
			t.Error("snapshot was deleted despite local delete failing")
		}
	})
}
