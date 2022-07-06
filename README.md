# Flotta Developer CLI

This repo has the developer CLI for [project-flotta.io](https://github.com/project-flotta).

## Installation

### Requirements

- `go >= 1.17`

### Build

Build the project by running: `make build`

## Usage

Use the developer CLI commands by running:

`./bin/flotta <command> <subcommand>`

For example:
`./bin/flotta add device`


```
Available Commands:
  add         Add a new flotta resource
  completion  Generate the autocompletion script for the specified shell
  delete      Delete a flotta resource
  help        Help about any command
  list        list flotta resources
  start       Start a flotta resource
  stop        Stop a flotta resource

Flags:
      --config string   config file (default is $HOME/.flotta-dev-cli.yaml)
  -h, --help            help for flotta
  -t, --toggle          Help message for toggle

Use "flotta [command] --help" for more information about a command.
```
