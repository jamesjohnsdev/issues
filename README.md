# Issues

This is CLI utility to facilitate easy creation, review and iteration of issues.

Essentially, this uses the [GitHub CLI](https://cli.github.com/) to enable local synronisation of issues.

The aim is to facilitate issue driven development, and improve both manual workflows where coding by hand, but
also for AI augmented workflows, where agents can have a source for development.

## Commands

### Core

- `issues init` - creates a new `.issues` directory in the current directory.
- `issues list` - lists all issues in the current directory.
- `issues create <title>` - creates a new issue in the current directory.
- `issues create -e` - opens a blank issue in the editor without requiring a title upfront (discarded if saved with no title).
- `issues view <number>` - opens an issue by number in the default editor.
- `issues view <number> -c` - prints all comments on an issue to stdout.
- `issues comment <number>` - opens the editor to draft a new comment (saved locally until pushed).
- `issues close <number>` - marks an issue as closed.
- `issues merge <a> <b>` - closes issue `a` as a duplicate of issue `b`, adding cross-reference comments.
- `issues pull` - pulls all issues (and their comments) from GitHub.
- `issues pull <number>` - pulls a single issue and its comments.
- `issues push` - pushes all modified issues and any new local comment drafts to GitHub.
- `issues push <number>` - pushes a single issue and any new local comment drafts.
- `issues sync` - syncs all issues in the current directory using the GitHub CLI.
- `issues help` - shows help for the `issues` command.

### Comments

Each issue has a colocated `.comments.json` file (e.g. `19-add-comment-support.comments.json`) that is created automatically when pulling an issue. It contains all comments fetched from GitHub.

**Typical workflow:**

```sh
issues pull 19                # fetch issue and its comments
issues view 19 -c             # read existing comments
issues comment 19             # open editor to write a new comment
issues push 19                # post the draft to GitHub
```

Comments with no `id` in the JSON file are treated as local drafts and posted to GitHub on the next `push`. After pushing, the file is updated with the assigned IDs.

### Agentic

These are planned commnads. The idea is that you will define a preferred agent in a config file.

- `issues agent` - sends the current issue to the agent.
- `issues agent list` - sends the list of issues to the agent.

There will be a agent skill created as as well down the line.

## How it works

The `.issues` directory is stored in the project root. It contains a `config.yaml` file. It contains an `issues` directory, which contains current local versions of the issues.

There is a `.remote_issues` directory, which contains the remote versions of the issues. During synchronisation, the local versions are updated to match the remote versions.
