package gh

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/jamesjohnsdev/issues/internal/issue"
)

const jsonFields = "number,title,state,stateReason,body,labels,assignees,milestone"
const commentFields = "comments"

type ghLabel struct {
	Name string `json:"name"`
}

type ghUser struct {
	Login string `json:"login"`
}

type ghMilestone struct {
	Title string `json:"title"`
}

type ghIssue struct {
	Number      int          `json:"number"`
	Title       string       `json:"title"`
	State       string       `json:"state"`
	StateReason string       `json:"stateReason"`
	Body        string       `json:"body"`
	Labels      []ghLabel    `json:"labels"`
	Assignees   []ghUser     `json:"assignees"`
	Milestone   *ghMilestone `json:"milestone"`
}

func (g ghIssue) toIssue() *issue.Issue {
	labels := make([]string, len(g.Labels))
	for i, l := range g.Labels {
		labels[i] = l.Name
	}
	assignees := make([]string, len(g.Assignees))
	for i, a := range g.Assignees {
		assignees[i] = a.Login
	}
	milestone := ""
	if g.Milestone != nil {
		milestone = g.Milestone.Title
	}
	return &issue.Issue{
		Number:      g.Number,
		Title:       g.Title,
		Labels:      labels,
		Assignees:   assignees,
		Milestone:   milestone,
		State:       strings.ToLower(g.State),
		StateReason: strings.ToLower(g.StateReason),
		Body:        g.Body,
	}
}

func ListAll() ([]*issue.Issue, error) {
	out, err := run("gh", "issue", "list",
		"--state", "all",
		"--limit", "1000",
		"--json", jsonFields,
	)
	if err != nil {
		return nil, fmt.Errorf("gh issue list: %w", err)
	}
	var raw []ghIssue
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, err
	}
	issues := make([]*issue.Issue, len(raw))
	for i, r := range raw {
		issues[i] = r.toIssue()
	}
	return issues, nil
}

func Get(number int) (*issue.Issue, error) {
	out, err := run("gh", "issue", "view", fmt.Sprintf("%d", number),
		"--json", jsonFields,
	)
	if err != nil {
		return nil, fmt.Errorf("gh issue view %d: %w", number, err)
	}
	var raw ghIssue
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, err
	}
	return raw.toIssue(), nil
}

// Create creates a new issue on GitHub and returns the assigned number.
func Create(iss *issue.Issue) (int, error) {
	args := []string{"issue", "create", "--title", iss.Title, "--body", iss.Body}
	for _, l := range iss.Labels {
		args = append(args, "--label", l)
	}
	for _, a := range iss.Assignees {
		args = append(args, "--assignee", a)
	}
	if iss.Milestone != "" {
		args = append(args, "--milestone", iss.Milestone)
	}

	out, err := run("gh", args...)
	if err != nil {
		return 0, fmt.Errorf("gh issue create: %w", err)
	}

	// output is the issue URL; parse number from last path segment
	url := strings.TrimSpace(string(out))
	parts := strings.Split(url, "/")
	var number int
	if _, err := fmt.Sscanf(parts[len(parts)-1], "%d", &number); err != nil || number == 0 {
		return 0, fmt.Errorf("could not parse issue number from: %s", url)
	}
	return number, nil
}

// Update pushes local edits to an existing GitHub issue.
func Update(iss *issue.Issue) error {
	current, err := Get(iss.Number)
	if err != nil {
		return err
	}

	args := []string{"issue", "edit", fmt.Sprintf("%d", iss.Number),
		"--title", iss.Title,
		"--body", iss.Body,
	}

	curLabels := toSet(current.Labels)
	newLabels := toSet(iss.Labels)
	for l := range newLabels {
		if !curLabels[l] {
			args = append(args, "--add-label", l)
		}
	}
	for l := range curLabels {
		if !newLabels[l] {
			args = append(args, "--remove-label", l)
		}
	}

	curAssignees := toSet(current.Assignees)
	newAssignees := toSet(iss.Assignees)
	for a := range newAssignees {
		if !curAssignees[a] {
			args = append(args, "--add-assignee", a)
		}
	}
	for a := range curAssignees {
		if !newAssignees[a] {
			args = append(args, "--remove-assignee", a)
		}
	}

	if iss.Milestone != "" {
		args = append(args, "--milestone", iss.Milestone)
	}

	if _, err := run("gh", args...); err != nil {
		return fmt.Errorf("gh issue edit %d: %w", iss.Number, err)
	}
	return nil
}

type ghComment struct {
	ID        string    `json:"id"`
	Author    ghUser    `json:"author"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"createdAt"`
}

// GetComments fetches all comments for a GitHub issue.
func GetComments(number int) ([]*issue.Comment, error) {
	out, err := run("gh", "issue", "view", fmt.Sprintf("%d", number),
		"--json", commentFields,
	)
	if err != nil {
		return nil, fmt.Errorf("gh issue view %d: %w", number, err)
	}
	var raw struct {
		Comments []ghComment `json:"comments"`
	}
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, err
	}
	comments := make([]*issue.Comment, len(raw.Comments))
	for i, c := range raw.Comments {
		t := c.CreatedAt
		comments[i] = &issue.Comment{
			ID:        c.ID,
			Author:    c.Author.Login,
			CreatedAt: &t,
			Body:      c.Body,
		}
	}
	return comments, nil
}

// AddComment posts a new comment on a GitHub issue.
func AddComment(number int, body string) error {
	if _, err := run("gh", "issue", "comment", fmt.Sprintf("%d", number), "--body", body); err != nil {
		return fmt.Errorf("gh issue comment %d: %w", number, err)
	}
	return nil
}

// Delete permanently deletes a GitHub issue.
func Delete(number int) error {
	if _, err := run("gh", "issue", "delete", fmt.Sprintf("%d", number), "--yes"); err != nil {
		return fmt.Errorf("gh issue delete %d: %w", number, err)
	}
	return nil
}

func Now() *time.Time {
	t := time.Now().UTC().Truncate(time.Second)
	return &t
}

func run(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("%w\n%s", err, string(ee.Stderr))
		}
	}
	return out, err
}

func toSet(ss []string) map[string]bool {
	m := make(map[string]bool, len(ss))
	for _, s := range ss {
		m[s] = true
	}
	return m
}
