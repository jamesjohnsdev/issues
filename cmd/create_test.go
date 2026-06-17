package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

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
