package issue

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Issue struct {
	Number      int        `yaml:"number,omitempty"`
	Title       string     `yaml:"title"`
	Labels      []string   `yaml:"labels,omitempty"`
	Assignees   []string   `yaml:"assignees,omitempty"`
	Milestone   string     `yaml:"milestone,omitempty"`
	State       string     `yaml:"state"`
	StateReason string     `yaml:"state_reason,omitempty"`
	SyncedAt    *time.Time `yaml:"synced_at,omitempty"`

	Body string `yaml:"-"`
	Path string `yaml:"-"`
}

func Parse(path string) (*Issue, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	s := string(data)
	if !strings.HasPrefix(s, "---\n") {
		return nil, fmt.Errorf("%s: missing frontmatter", path)
	}
	rest := s[4:]
	end := strings.Index(rest, "\n---")
	if end == -1 {
		return nil, fmt.Errorf("%s: unclosed frontmatter", path)
	}
	var iss Issue
	if err := yaml.Unmarshal([]byte(rest[:end]), &iss); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	iss.Body = strings.TrimLeft(rest[end+4:], "\n")
	iss.Path = path
	return &iss, nil
}

func Write(path string, iss *Issue) error {
	var buf bytes.Buffer
	buf.WriteString("---\n")
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(iss); err != nil {
		return err
	}
	_ = enc.Close()
	buf.WriteString("---\n")
	if iss.Body != "" {
		buf.WriteString("\n")
		body := iss.Body
		if !strings.HasSuffix(body, "\n") {
			body += "\n"
		}
		buf.WriteString(body)
	}
	return os.WriteFile(path, buf.Bytes(), 0644)
}

// WriteTemplate writes an issue file with all user-editable fields expanded
// (even when empty) and a comment marking where the body begins. Used when
// opening a new issue in the editor so the user sees the full schema.
func WriteTemplate(path string, iss *Issue) error {
	type tpl struct {
		Title     string   `yaml:"title"`
		State     string   `yaml:"state"`
		Labels    []string `yaml:"labels"`
		Assignees []string `yaml:"assignees"`
		Milestone string   `yaml:"milestone"`
	}
	labels := iss.Labels
	if labels == nil {
		labels = []string{}
	}
	assignees := iss.Assignees
	if assignees == nil {
		assignees = []string{}
	}
	t := tpl{
		Title:     iss.Title,
		State:     iss.State,
		Labels:    labels,
		Assignees: assignees,
		Milestone: iss.Milestone,
	}
	var buf bytes.Buffer
	buf.WriteString("---\n")
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(t); err != nil {
		return err
	}
	_ = enc.Close()
	buf.WriteString("# Body goes below this line\n")
	buf.WriteString("---\n")
	if iss.Body != "" {
		buf.WriteString("\n")
		body := iss.Body
		if !strings.HasSuffix(body, "\n") {
			body += "\n"
		}
		buf.WriteString(body)
	}
	return os.WriteFile(path, buf.Bytes(), 0644)
}

func Filename(iss *Issue) string {
	return fmt.Sprintf("%d-%s.md", iss.Number, Slug(iss.Title))
}
