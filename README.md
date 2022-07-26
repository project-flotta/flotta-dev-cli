# Flotta Developer CLI

This repo has the developer CLI for [project-flotta.io](https://github.com/project-flotta).
The purpose of this cli is easily create edge devices and deploy prefefine workloads on.

## Installation
Installing flotta-dev-cli from rpm and its usage are describe on project's [blog-post](https://project-flotta.io/flotta/2022/07/20/developer-cli.html).

## Development
### Requirements

- go >= 1.17
- docker

### Build

Build the project by running: `make build`

## Usage

Use the developer CLI commands by running:

`./bin/flotta <command> <subcommand>`

For example:
`bin/flotta add device --name mydevice`


```
Usage:
  flotta [command]

Available Commands:
  add         Add a new flotta resource
  completion  Generate the autocompletion script for the specified shell
  delete      Delete the flotta resource
  help        Help about any command
  list        List flotta resources
  start       Start flotta resource
  stop        Stop flotta resource

Flags:
  -h, --help   help for flotta

Use "flotta [command] --help" for more information about a command.
```
