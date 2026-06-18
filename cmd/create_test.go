package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesjohnsdev/issues/internal/issue"
)

func resetCreateFlag(t *testing.T) {
	t.Helper()
	orig := createEditorFlag
	t.Cleanup(func() { createEditorFlag = orig })
	createEditorFlag = false
}

func openIssuesDir(t *testing.T, parent string) string {
	t.Helper()
	return filepath.Join(parent, issuesDirName, "open")
}

func TestCreateInteractive(t *testing.T) {
	t.Run("empty title aborts with no file", func(t *testing.T) {
		parent := makeProjectDir(t, nil)
		chdirTo(t, parent)
		resetCreateFlag(t)
		t.Setenv("VISUAL", "")
		t.Setenv("EDITOR", "")

		injectStdin(t, "\n\n\n\n\n")
		_ = captureStdout(t, func() {
			if err := createCmd.RunE(createCmd, nil); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})

		if files := readMDFiles(t, openIssuesDir(t, parent)); len(files) != 0 {
			t.Errorf("expected no files after empty title, got %d", len(files))
		}
	})

	t.Run("whitespace-only title aborts", func(t *testing.T) {
		parent := makeProjectDir(t, nil)
		chdirTo(t, parent)
		resetCreateFlag(t)
		t.Setenv("VISUAL", "")
		t.Setenv("EDITOR", "")

		injectStdin(t, "   \n\n\n\n\n")
		_ = captureStdout(t, func() {
			if err := createCmd.RunE(createCmd, nil); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})

		if files := readMDFiles(t, openIssuesDir(t, parent)); len(files) != 0 {
			t.Errorf("expected no files after whitespace title, got %d", len(files))
		}
	})

	t.Run("title only creates file", func(t *testing.T) {
		parent := makeProjectDir(t, nil)
		chdirTo(t, parent)
		resetCreateFlag(t)
		t.Setenv("VISUAL", "")
		t.Setenv("EDITOR", "")

		injectStdin(t, "My Bug\n\n\n\n\n")
		_ = captureStdout(t, func() {
			if err := createCmd.RunE(createCmd, nil); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})

		files := readMDFiles(t, openIssuesDir(t, parent))
		if len(files) != 1 {
			t.Fatalf("expected 1 file, got %d", len(files))
		}
		iss, err := issue.Parse(files[0])
		if err != nil {
			t.Fatalf("parse error: %v", err)
		}
		if iss.Title != "My Bug" {
			t.Errorf("Title = %q, want %q", iss.Title, "My Bug")
		}
		if iss.State != "open" {
			t.Errorf("State = %q, want %q", iss.State, "open")
		}
		if !strings.Contains(filepath.Base(files[0]), "my-bug") {
			t.Errorf("filename %q should contain slug %q", filepath.Base(files[0]), "my-bug")
		}
	})

	t.Run("labels parsed and written", func(t *testing.T) {
		parent := makeProjectDir(t, nil)
		chdirTo(t, parent)
		resetCreateFlag(t)
		t.Setenv("VISUAL", "")
		t.Setenv("EDITOR", "")

		injectStdin(t, "Fix Auth\n\nbug, auth\n\n\n")
		_ = captureStdout(t, func() {
			if err := createCmd.RunE(createCmd, nil); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})

		files := readMDFiles(t, openIssuesDir(t, parent))
		if len(files) != 1 {
			t.Fatalf("expected 1 file, got %d", len(files))
		}
		iss, err := issue.Parse(files[0])
		if err != nil {
			t.Fatalf("parse error: %v", err)
		}
		want := []string{"bug", "auth"}
		if len(iss.Labels) != len(want) {
			t.Fatalf("Labels = %v, want %v", iss.Labels, want)
		}
		for i, l := range want {
			if iss.Labels[i] != l {
				t.Errorf("Labels[%d] = %q, want %q", i, iss.Labels[i], l)
			}
		}
	})

	t.Run("assignees parsed and written", func(t *testing.T) {
		parent := makeProjectDir(t, nil)
		chdirTo(t, parent)
		resetCreateFlag(t)
		t.Setenv("VISUAL", "")
		t.Setenv("EDITOR", "")

		injectStdin(t, "Fix Auth\n\n\njames, alice\n\n")
		_ = captureStdout(t, func() {
			if err := createCmd.RunE(createCmd, nil); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})

		files := readMDFiles(t, openIssuesDir(t, parent))
		if len(files) != 1 {
			t.Fatalf("expected 1 file, got %d", len(files))
		}
		iss, err := issue.Parse(files[0])
		if err != nil {
			t.Fatalf("parse error: %v", err)
		}
		want := []string{"james", "alice"}
		if len(iss.Assignees) != len(want) {
			t.Fatalf("Assignees = %v, want %v", iss.Assignees, want)
		}
		for i, a := range want {
			if iss.Assignees[i] != a {
				t.Errorf("Assignees[%d] = %q, want %q", i, iss.Assignees[i], a)
			}
		}
	})

	t.Run("milestone set", func(t *testing.T) {
		parent := makeProjectDir(t, nil)
		chdirTo(t, parent)
		resetCreateFlag(t)
		t.Setenv("VISUAL", "")
		t.Setenv("EDITOR", "")

		injectStdin(t, "Fix Auth\n\n\n\nv1.2\n")
		_ = captureStdout(t, func() {
			if err := createCmd.RunE(createCmd, nil); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})

		files := readMDFiles(t, openIssuesDir(t, parent))
		if len(files) != 1 {
			t.Fatalf("expected 1 file, got %d", len(files))
		}
		iss, err := issue.Parse(files[0])
		if err != nil {
			t.Fatalf("parse error: %v", err)
		}
		if iss.Milestone != "v1.2" {
			t.Errorf("Milestone = %q, want %q", iss.Milestone, "v1.2")
		}
	})

	t.Run("all metadata fields", func(t *testing.T) {
		parent := makeProjectDir(t, nil)
		chdirTo(t, parent)
		resetCreateFlag(t)
		t.Setenv("VISUAL", "")
		t.Setenv("EDITOR", "")

		injectStdin(t, "Full Issue\n\nbug\njames\nv2.0\n")
		_ = captureStdout(t, func() {
			if err := createCmd.RunE(createCmd, nil); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})

		files := readMDFiles(t, openIssuesDir(t, parent))
		if len(files) != 1 {
			t.Fatalf("expected 1 file, got %d", len(files))
		}
		iss, err := issue.Parse(files[0])
		if err != nil {
			t.Fatalf("parse error: %v", err)
		}
		if iss.Title != "Full Issue" {
			t.Errorf("Title = %q, want %q", iss.Title, "Full Issue")
		}
		if len(iss.Labels) != 1 || iss.Labels[0] != "bug" {
			t.Errorf("Labels = %v, want [bug]", iss.Labels)
		}
		if len(iss.Assignees) != 1 || iss.Assignees[0] != "james" {
			t.Errorf("Assignees = %v, want [james]", iss.Assignees)
		}
		if iss.Milestone != "v2.0" {
			t.Errorf("Milestone = %q, want %q", iss.Milestone, "v2.0")
		}
	})

	t.Run("body e with no editor skips and still creates file", func(t *testing.T) {
		parent := makeProjectDir(t, nil)
		chdirTo(t, parent)
		resetCreateFlag(t)
		t.Setenv("VISUAL", "")
		t.Setenv("EDITOR", "")

		injectStdin(t, "My Bug\ne\n\n\n\n")
		_ = captureStdout(t, func() {
			if err := createCmd.RunE(createCmd, nil); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})

		if files := readMDFiles(t, openIssuesDir(t, parent)); len(files) != 1 {
			t.Errorf("expected 1 file, got %d", len(files))
		}
	})

	t.Run("body e with editor writes body", func(t *testing.T) {
		parent := makeProjectDir(t, nil)
		chdirTo(t, parent)
		resetCreateFlag(t)

		script := filepath.Join(t.TempDir(), "editor.sh")
		if err := os.WriteFile(script, []byte(
			"#!/bin/sh\ncat > \"$1\" <<'EOF'\n---\ntitle: My Bug\nstate: open\n---\n\nsome body\nEOF\n",
		), 0755); err != nil {
			t.Fatal(err)
		}
		t.Setenv("VISUAL", "")
		t.Setenv("EDITOR", script)

		// Only title + body "e" — no labels/assignees/milestone prompts should appear.
		injectStdin(t, "My Bug\ne\n")
		_ = captureStdout(t, func() {
			if err := createCmd.RunE(createCmd, nil); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})

		files := readMDFiles(t, openIssuesDir(t, parent))
		if len(files) != 1 {
			t.Fatalf("expected 1 file, got %d", len(files))
		}
		iss, err := issue.Parse(files[0])
		if err != nil {
			t.Fatalf("parse error: %v", err)
		}
		if iss.Body == "" {
			t.Error("expected non-empty body after editor wrote content")
		}
	})

	t.Run("body e with editor skips labels assignees milestone prompts", func(t *testing.T) {
		parent := makeProjectDir(t, nil)
		chdirTo(t, parent)
		resetCreateFlag(t)

		script := filepath.Join(t.TempDir(), "editor.sh")
		if err := os.WriteFile(script, []byte(
			"#!/bin/sh\ncat > \"$1\" <<'EOF'\n---\ntitle: My Bug\nstate: open\n---\nEOF\n",
		), 0755); err != nil {
			t.Fatal(err)
		}
		t.Setenv("VISUAL", "")
		t.Setenv("EDITOR", script)

		// If Labels/Assignees/Milestone prompts were shown, this input would
		// block waiting for three more lines. Completing without hanging
		// confirms they are skipped.
		injectStdin(t, "My Bug\ne\n")
		_ = captureStdout(t, func() {
			if err := createCmd.RunE(createCmd, nil); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})

		if files := readMDFiles(t, openIssuesDir(t, parent)); len(files) != 1 {
			t.Errorf("expected 1 file, got %d", len(files))
		}
	})

	t.Run("T-number increments from existing issues", func(t *testing.T) {
		parent := makeProjectDir(t, []issueFixture{
			{"T1-existing.md", issue.Issue{Title: "Existing", State: "open"}},
			{"T2-another.md", issue.Issue{Title: "Another", State: "open"}},
		})
		chdirTo(t, parent)
		resetCreateFlag(t)
		t.Setenv("VISUAL", "")
		t.Setenv("EDITOR", "")

		injectStdin(t, "New Issue\n\n\n\n\n")
		_ = captureStdout(t, func() {
			if err := createCmd.RunE(createCmd, nil); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})

		files := readMDFiles(t, openIssuesDir(t, parent))
		var newFile string
		for _, f := range files {
			base := filepath.Base(f)
			if strings.HasPrefix(base, "T3-") {
				newFile = f
			}
		}
		if newFile == "" {
			t.Errorf("expected a file prefixed T3-, got files: %v", files)
		}
	})
}

func TestCreateEditorFlag(t *testing.T) {
	t.Run("no editor set returns error", func(t *testing.T) {
		parent := makeProjectDir(t, nil)
		chdirTo(t, parent)
		t.Setenv("VISUAL", "")
		t.Setenv("EDITOR", "")

		orig := createEditorFlag
		t.Cleanup(func() { createEditorFlag = orig })
		createEditorFlag = true

		err := createCmd.RunE(createCmd, nil)
		if err == nil {
			t.Error("expected error when no editor is set, got nil")
		}
	})

	t.Run("no-op editor deletes file with empty title", func(t *testing.T) {
		parent := makeProjectDir(t, nil)
		chdirTo(t, parent)
		t.Setenv("VISUAL", "")
		t.Setenv("EDITOR", "true")

		orig := createEditorFlag
		t.Cleanup(func() { createEditorFlag = orig })
		createEditorFlag = true

		_ = captureStdout(t, func() {
			if err := createCmd.RunE(createCmd, nil); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})

		entries, err := os.ReadDir(filepath.Join(parent, issuesDirName, "open"))
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range entries {
			if filepath.Ext(e.Name()) == ".md" {
				t.Errorf("expected no .md files after aborting, found %s", e.Name())
			}
		}
	})

	t.Run("editor that sets a title keeps the file", func(t *testing.T) {
		parent := makeProjectDir(t, nil)
		chdirTo(t, parent)

		script := filepath.Join(t.TempDir(), "editor.sh")
		if err := os.WriteFile(script, []byte(
			"#!/bin/sh\ncat > \"$1\" <<'EOF'\n---\ntitle: My New Issue\nstate: open\n---\nEOF\n",
		), 0755); err != nil {
			t.Fatal(err)
		}
		t.Setenv("VISUAL", "")
		t.Setenv("EDITOR", script)

		orig := createEditorFlag
		t.Cleanup(func() { createEditorFlag = orig })
		createEditorFlag = true

		_ = captureStdout(t, func() {
			if err := createCmd.RunE(createCmd, nil); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})

		entries, err := os.ReadDir(filepath.Join(parent, issuesDirName, "open"))
		if err != nil {
			t.Fatal(err)
		}
		var mdFiles []string
		for _, e := range entries {
			if filepath.Ext(e.Name()) == ".md" {
				mdFiles = append(mdFiles, e.Name())
			}
		}
		if len(mdFiles) != 1 {
			t.Errorf("expected 1 .md file, got %d: %v", len(mdFiles), mdFiles)
		}
	})
}
