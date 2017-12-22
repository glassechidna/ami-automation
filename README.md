[![Build Status](https://travis-ci.org/glassechidna/ami-automation.svg?branch=master)](https://travis-ci.org/glassechidna/ami-automation)

```
$ ./ami-automation

ami-automation is a CLI tool to make using AWS' SSM Automation functionality
from either a terminal or a CI system as easy as possible.

The 'start' subcommand will start an automation execution and stream its
progress to stderr until the automation has finished. The automation document's
outputs are printed to stdout on completion. A failure exit code will be returned
if the automation fails.

Usage:
  ami-automation [command]

Available Commands:
  help        Help about any command
  show        Show output of SSM automation that has already happened
  start       Start an SSM automation execution
  version     Output ami-automation version information

Flags:
  -h, --help   help for ami-automation

Use "ami-automation [command] --help" for more information about a command.
```
