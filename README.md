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
- `issues close <number>` - marks an issue as closed.
- `issues pull` - pulls all issues (and their comments) from GitHub.
- `issues pull <number>` - pulls a single issue and its comments.
- `issues push` - pushes all modified issues and any new local comments to GitHub.
- `issues push <number>` - pushes a single issue and any new local comments.
- `issues sync` - syncs all issues in the current directory using the GitHub CLI.
- `issues help` - shows help for the `issues` command.

### Comments

Each issue has a colocated `.comments.json` file (e.g. `19-add-comment-support.comments.json`) that is created automatically when pulling an issue. It contains all comments fetched from GitHub.

To add a new comment locally, append an entry with only a `body` field:

```json
[
  {
    "id": "IC_kwDO...",
    "author": "jamesjohnsdev",
    "created_at": "2026-06-20T10:00:00Z",
    "body": "Existing comment from GitHub."
  },
  {
    "body": "My new comment — will be posted on the next push."
  }
]
```

Running `issues push <number>` will post any entries without an `id` to GitHub and update the file with their assigned IDs.

### Agentic

These are planned commnads. The idea is that you will define a preferred agent in a config file.

- `issues agent` - sends the current issue to the agent.
- `issues agent list` - sends the list of issues to the agent.

There will be a agent skill created as as well down the line.

## How it works

The `.issues` directory is stored in the project root. It contains a `config.yaml` file. It contains an `issues` directory, which contains current local versions of the issues.

There is a `.remote_issues` directory, which contains the remote versions of the issues. During synchronisation, the local versions are updated to match the remote versions.
