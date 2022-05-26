# Contributing Guidelines

## Installation

### Requirements

- `go >= 1.17`
- cobra cli package: `go install github.com/spf13/cobra-cli@latest`

### Build

Build the project by running: `go build .`

## Contribute

### Add a new command

- To add a new command run: `cobra add <command name>`.

  This will create a new file named `<command name>.go` inside the `cmd` directory.

- To add a new subcommand run: `cobra add <subcommand name> -p '<command>Cmd'`.

  This will create a new file named `<subcommand name>.go` inside the `cmd` directory.

### Apply the changes
Once you have finished editing the new file, re-build the project: `go build .`.
