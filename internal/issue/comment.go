package issue

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Comment struct {
	ID        string     `json:"id,omitempty"`
	Author    string     `json:"author,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	Body      string     `json:"body"`
}

func CommentsFilename(iss *Issue) string {
	return fmt.Sprintf("%d-%s.comments.json", iss.Number, Slug(iss.Title))
}

func ParseComments(path string) ([]*Comment, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var comments []*Comment
	if err := json.Unmarshal(data, &comments); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return comments, nil
}

func WriteComments(path string, comments []*Comment) error {
	data, err := json.MarshalIndent(comments, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0644)
}
