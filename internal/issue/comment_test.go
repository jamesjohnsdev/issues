package issue

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCommentsFilename(t *testing.T) {
	tests := []struct {
		name string
		iss  Issue
		want string
	}{
		{
			name: "basic issue",
			iss:  Issue{Number: 1, Title: "Hello World"},
			want: "1-hello-world.comments.json",
		},
		{
			name: "special chars in title",
			iss:  Issue{Number: 42, Title: "Fix: bug #42!"},
			want: "42-fix-bug-42.comments.json",
		},
		{
			name: "long title is truncated",
			iss:  Issue{Number: 3, Title: "this is a very long title that goes way beyond fifty characters"},
			want: "3-this-is-a-very-long-title-that-goes-way-beyond.comments.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CommentsFilename(&tt.iss)
			if got != tt.want {
				t.Errorf("CommentsFilename() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseComments(t *testing.T) {
	t.Run("missing file returns nil slice", func(t *testing.T) {
		comments, err := ParseComments("/no/such/file.comments.json")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if comments != nil {
			t.Errorf("expected nil, got %v", comments)
		}
	})

	t.Run("empty array", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "test.comments.json")
		if err := os.WriteFile(path, []byte("[]\n"), 0644); err != nil {
			t.Fatal(err)
		}
		comments, err := ParseComments(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(comments) != 0 {
			t.Errorf("got %d comments, want 0", len(comments))
		}
	})

	t.Run("synced comment with all fields", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "test.comments.json")
		raw := `[{"metadata":{"id":"IC_abc123","author":"alice","created_at":"2026-01-01T00:00:00Z"},"body":"hello"}]`
		if err := os.WriteFile(path, []byte(raw), 0644); err != nil {
			t.Fatal(err)
		}
		comments, err := ParseComments(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(comments) != 1 {
			t.Fatalf("got %d comments, want 1", len(comments))
		}
		c := comments[0]
		if c.Metadata == nil {
			t.Fatal("Metadata is nil, want non-nil")
		}
		if c.Metadata.ID != "IC_abc123" {
			t.Errorf("ID = %q, want %q", c.Metadata.ID, "IC_abc123")
		}
		if c.Metadata.Author != "alice" {
			t.Errorf("Author = %q, want %q", c.Metadata.Author, "alice")
		}
		if c.Body != "hello" {
			t.Errorf("Body = %q, want %q", c.Body, "hello")
		}
	})

	t.Run("local-only comment with body only", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "test.comments.json")
		raw := `[{"body":"draft comment"}]`
		if err := os.WriteFile(path, []byte(raw), 0644); err != nil {
			t.Fatal(err)
		}
		comments, err := ParseComments(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(comments) != 1 {
			t.Fatalf("got %d comments, want 1", len(comments))
		}
		if comments[0].Metadata != nil {
			t.Errorf("expected nil Metadata for local comment, got %+v", comments[0].Metadata)
		}
		if comments[0].Body != "draft comment" {
			t.Errorf("Body = %q, want %q", comments[0].Body, "draft comment")
		}
	})

	t.Run("invalid JSON returns error", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "test.comments.json")
		if err := os.WriteFile(path, []byte("not json"), 0644); err != nil {
			t.Fatal(err)
		}
		_, err := ParseComments(path)
		if err == nil {
			t.Error("expected error for invalid JSON, got nil")
		}
	})
}

func TestWriteParseComments(t *testing.T) {
	t.Run("empty slice round-trips", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "test.comments.json")
		if err := WriteComments(path, []*Comment{}); err != nil {
			t.Fatalf("WriteComments: %v", err)
		}
		got, err := ParseComments(path)
		if err != nil {
			t.Fatalf("ParseComments: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("got %d comments, want 0", len(got))
		}
	})

	t.Run("synced comment round-trips", func(t *testing.T) {
		ts := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		in := []*Comment{
			{Metadata: &CommentMeta{ID: "IC_abc", Author: "alice", CreatedAt: &ts}, Body: "hello"},
		}
		path := filepath.Join(t.TempDir(), "test.comments.json")
		if err := WriteComments(path, in); err != nil {
			t.Fatalf("WriteComments: %v", err)
		}
		got, err := ParseComments(path)
		if err != nil {
			t.Fatalf("ParseComments: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("got %d comments, want 1", len(got))
		}
		c := got[0]
		if c.Metadata == nil {
			t.Fatal("Metadata is nil, want non-nil")
		}
		if c.Metadata.ID != "IC_abc" {
			t.Errorf("ID = %q, want %q", c.Metadata.ID, "IC_abc")
		}
		if c.Metadata.Author != "alice" {
			t.Errorf("Author = %q, want %q", c.Metadata.Author, "alice")
		}
		if c.Body != "hello" {
			t.Errorf("Body = %q, want %q", c.Body, "hello")
		}
		if !c.Metadata.CreatedAt.Equal(ts) {
			t.Errorf("CreatedAt = %v, want %v", c.Metadata.CreatedAt, ts)
		}
	})

	t.Run("local-only comment round-trips with nil metadata", func(t *testing.T) {
		in := []*Comment{{Body: "new comment"}}
		path := filepath.Join(t.TempDir(), "test.comments.json")
		if err := WriteComments(path, in); err != nil {
			t.Fatalf("WriteComments: %v", err)
		}
		got, err := ParseComments(path)
		if err != nil {
			t.Fatalf("ParseComments: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("got %d comments, want 1", len(got))
		}
		if got[0].Metadata != nil {
			t.Errorf("expected nil Metadata, got %+v", got[0].Metadata)
		}
		if got[0].Body != "new comment" {
			t.Errorf("Body = %q, want %q", got[0].Body, "new comment")
		}
	})

	t.Run("mixed synced and local comments round-trip", func(t *testing.T) {
		ts := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
		in := []*Comment{
			{Metadata: &CommentMeta{ID: "IC_xyz", Author: "bob", CreatedAt: &ts}, Body: "existing"},
			{Body: "my draft"},
		}
		path := filepath.Join(t.TempDir(), "test.comments.json")
		if err := WriteComments(path, in); err != nil {
			t.Fatalf("WriteComments: %v", err)
		}
		got, err := ParseComments(path)
		if err != nil {
			t.Fatalf("ParseComments: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("got %d comments, want 2", len(got))
		}
		if got[0].Metadata == nil || got[0].Metadata.ID != "IC_xyz" {
			t.Errorf("got[0].Metadata.ID = %v, want %q", got[0].Metadata, "IC_xyz")
		}
		if got[1].Metadata != nil {
			t.Errorf("got[1].Metadata = %+v, want nil", got[1].Metadata)
		}
		if got[1].Body != "my draft" {
			t.Errorf("got[1].Body = %q, want %q", got[1].Body, "my draft")
		}
	})

	t.Run("output file ends with newline", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "test.comments.json")
		if err := WriteComments(path, []*Comment{{Body: "x"}}); err != nil {
			t.Fatalf("WriteComments: %v", err)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if len(data) == 0 || data[len(data)-1] != '\n' {
			t.Errorf("file does not end with newline: %q", string(data))
		}
	})
}
