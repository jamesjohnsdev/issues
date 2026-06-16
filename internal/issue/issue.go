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

func Filename(iss *Issue) string {
	return fmt.Sprintf("%d-%s.md", iss.Number, Slug(iss.Title))
}
