# Issues

This is CLI utility to facilitate easy creation, review and iteration of issues.

Essentially, this uses the [GitHub CLI](https://cli.github.com/) to enable local synronisation of issues.

The aim is to facilitate issue driven development, and improve both manual workflows where coding by hand, but
also for AI augmented workflows, where agents can have a source for development.

## Commands

### Core

- `issues init` - creates a new `.issues` directory in the current directory.
- `issues list` - lists all issues in the current directory.
- `issues create` - creates a new issue in the current directory.
- `issues view <number>` - opens an issue by number in the default editor.
- `issues sync` - syncs all issues in the current directory using the GitHub CLI.
- `issues help` - shows help for the `issues` command.

### Agentic

These are planned commnads. The idea is that you will define a preferred agent in a config file.

- `issues agent` - sends the current issue to the agent.
- `issues agent list` - sends the list of issues to the agent.

There will be a agent skill created as as well down the line.

## How it works

The `.issues` directory is stored in the project root. It contains a `config.yaml` file. It contains an `issues` directory, which contains current local versions of the issues.

There is a `.remote_issues` directory, which contains the remote versions of the issues. During synchronisation, the local versions are updated to match the remote versions.
