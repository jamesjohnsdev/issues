package issue

import "testing"

func TestSlug(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "basic words",
			input: "Hello World",
			want:  "hello-world",
		},
		{
			name:  "already lowercase",
			input: "hello world",
			want:  "hello-world",
		},
		{
			name:  "special characters stripped",
			input: "Fix: bug #42!",
			want:  "fix-bug-42",
		},
		{
			name:  "consecutive special chars collapsed to single dash",
			input: "hello   world",
			want:  "hello-world",
		},
		{
			name:  "leading and trailing special chars trimmed",
			input: "---hello world---",
			want:  "hello-world",
		},
		{
			name:  "numbers preserved",
			input: "issue 123 fix",
			want:  "issue-123-fix",
		},
		{
			name:  "exactly 50 chars not truncated",
			input: "enable flag passing to filter status on issue list",
			want:  "enable-flag-passing-to-filter-status-on-issue-list",
		},
		{
			name:  "over 50 chars truncated at word boundary",
			input: "this is a very long title that goes way beyond fifty characters",
			want:  "this-is-a-very-long-title-that-goes-way-beyond",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "only special chars",
			input: "!!! ???",
			want:  "",
		},
		{
			name:  "mixed alphanumeric and special",
			input: "feat(auth): add OAuth2 support",
			want:  "feat-auth-add-oauth2-support",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Slug(tt.input)
			if got != tt.want {
				t.Errorf("Slug(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
