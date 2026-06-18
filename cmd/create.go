package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/jamesjohnsdev/issues/internal/issue"
	"github.com/spf13/cobra"
)

var createEditorFlag bool

var createCmd = &cobra.Command{
	Use:   "create [title]",
	Short: "Create a new local issue",
	Args: func(cmd *cobra.Command, args []string) error {
		if createEditorFlag {
			return cobra.NoArgs(cmd, args)
		}
		if len(args) > 1 {
			return fmt.Errorf("accepts at most 1 arg, received %d", len(args))
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := issuesRoot()
		if err != nil {
			return err
		}

		existing, err := loadAllLocal(root)
		if err != nil {
			return err
		}
		maxT := 0
		for _, e := range existing {
			if e.Number == 0 {
				var n int
				if _, err := fmt.Sscanf(idFromPath(e.Path), "T%d", &n); err != nil {
					return err
				}
				if n > maxT {
					maxT = n
				}
			}
		}

		editor := os.Getenv("VISUAL")
		if editor == "" {
			editor = os.Getenv("EDITOR")
		}

		interactive := !createEditorFlag && len(args) == 0

		var filename string
		var openEditor bool
		iss := &issue.Issue{State: "open"}

		switch {
		case createEditorFlag:
			filename = fmt.Sprintf("T%d-new-issue.md", maxT+1)
			openEditor = true
			if editor == "" {
				return fmt.Errorf("no editor set: define $VISUAL or $EDITOR")
			}

		case interactive:
			reader := bufio.NewReader(os.Stdin)

			prompt := func(label string) (string, error) {
				fmt.Print(label)
				line, err := reader.ReadString('\n')
				return strings.TrimSpace(line), err
			}

			iss.Title, err = prompt("Title: ")
			if err != nil {
				return err
			}
			if iss.Title == "" {
				fmt.Println(color.YellowString("Aborted.") + " No title, issue discarded.")
				return nil
			}
			filename = fmt.Sprintf("T%d-%s.md", maxT+1, issue.Slug(iss.Title))

			var bodyPrompt string
			if editor != "" {
				bodyPrompt = fmt.Sprintf("Body [(e) to launch %s, Enter to skip]: ", filepath.Base(editor))
			} else {
				bodyPrompt = "Body [Enter to skip]: "
			}
			bodyResp, err := prompt(bodyPrompt)
			if err != nil {
				return err
			}
			openEditor = strings.ToLower(bodyResp) == "e"
			if openEditor && editor == "" {
				fmt.Println(color.YellowString("No editor set") + " ($VISUAL/$EDITOR). Skipping body.")
				openEditor = false
			}

			if !openEditor {
				labelsResp, err := prompt("Labels (comma-separated, Enter to skip): ")
				if err != nil {
					return err
				}
				for _, l := range strings.Split(labelsResp, ",") {
					if l := strings.TrimSpace(l); l != "" {
						iss.Labels = append(iss.Labels, l)
					}
				}

				assigneesResp, err := prompt("Assignees (comma-separated, Enter to skip): ")
				if err != nil {
					return err
				}
				for _, a := range strings.Split(assigneesResp, ",") {
					if a := strings.TrimSpace(a); a != "" {
						iss.Assignees = append(iss.Assignees, a)
					}
				}

				iss.Milestone, err = prompt("Milestone (Enter to skip): ")
				if err != nil {
					return err
				}
			}

		default:
			iss.Title = args[0]
			filename = fmt.Sprintf("T%d-%s.md", maxT+1, issue.Slug(iss.Title))
			openEditor = editor != ""
		}

		path := filepath.Join(openDir(root), filename)

		var writeErr error
		if openEditor {
			writeErr = issue.WriteTemplate(path, iss)
		} else {
			writeErr = issue.Write(path, iss)
		}
		if writeErr != nil {
			return writeErr
		}

		if openEditor {
			parts := strings.Fields(editor)
			c := exec.Command(parts[0], append(parts[1:], path)...)
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			if err := c.Run(); err != nil {
				return fmt.Errorf("editor exited with error: %w", err)
			}
		}

		if createEditorFlag {
			saved, err := issue.Parse(path)
			if err != nil || saved.Title == "" {
				_ = os.Remove(path)
				fmt.Println(color.YellowString("Aborted.") + " No title set, issue discarded.")
				return nil
			}
		}

		fmt.Printf("%s %s\n", color.GreenString("Created"), path)
		return nil
	},
}

func init() {
	createCmd.Flags().BoolVarP(&createEditorFlag, "editor", "e", false, "open a new blank issue directly in the editor")
}
