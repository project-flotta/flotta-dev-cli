# Contributing Guidelines

## Installation

### Requirements

- `go >= 1.17`
- cobra cli package: `go install github.com/spf13/cobra-cli@latest`

### Build

Build the project by running: `make build`

## Contribute

### Add a new command

In order to use `cobra-cli` and generate a new command, change your working path to the `internal` directory. 

#### Parent command

1. Create a new directory with the name of the parent command.

    For example, in order to add the command `add edgedevice`, create a new directory under `internal/` named `add`.


2. Import the new directory in the `main.go` file, by adding the next line to the import section:
   
    `_ "github.com/arielireni/flotta-dev-cli/internal/cmd/<parent command name>"`


3. Use the cobra generator to create the new command:

   `cobra add <command name>`

     This will create a new file named `<parent command name>.go` inside the `cmd` directory.


4. Move the new file to the parent command directory.

#### Sub-command

1. Use the cobra generator to create the new command:

    `cobra add <sub-command name> -p '<command>Cmd'`

      This will create a new file named `<sub-command name>.go` inside the `cmd` directory.


2. Move the new file to the parent command directory.

### Edit

#### New command

1. Change `rootCmd` to public by renaming it to `RootCmd`, so it will be importable in the `<parent command name>.go` file.


2. Make your own changes in the new generated file.

#### Existing command

- To edit parent command, edit the file `<parent command>.go` in `/internal/cmd/<parent command name>`. 

- To edit sub-command, edit the file `<sub-command name>.go` in `/internal/cmd/<parent command name>`.

### Apply the changes

Once you have finished editing the files, re-build the project: `make build`.
