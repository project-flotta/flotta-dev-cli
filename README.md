# Flotta Developer CLI

This repo has the developer CLI for [project-flotta.io](https://github.com/project-flotta).

## Installation

### Requirements

- `go >= 1.17`
- cobra package: `go install github.com/spf13/cobra-cli@latest`

## Usage

### Run existing command
`./flotta-dev-cli <command> <subcommand>`

For example:
`./flotta-dev-cli add edgedevice`

### Add a new command
- To add a new command run: `cobra add <command name>`. 

  This will create a new file named `<command name>.go` inside the `cmd` directory.

- To add a new subcommand run: `cobra add <subcommand name> -p '<command>Cmd'`. 

  This will create a new file named `<subcommand name>.go` inside the `cmd` directory.

Edit the new file as you wish and re-build the project: `go build .`