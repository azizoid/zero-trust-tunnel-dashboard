# Scripts Directory

This directory contains modular scripts for the `run` command.

## Structure

```
scripts/
└── lib/
    ├── commands.sh    # Command validation and help functions
    ├── ssh.sh         # SSH configuration functions
    ├── build.sh       # Build functions
    └── interactive.sh # Interactive connection setup
```

## Modules

### `commands.sh`
- `is_valid_command()` - Validates if a command is supported
- `get_command_desc()` - Returns description for a command
- `show_commands()` - Displays available commands and examples

### `ssh.sh`
- `list_ssh_hosts()` - Lists available hosts from `~/.ssh/config`

### `build.sh`
- `needs_build()` - Checks if binary needs to be rebuilt
- `build_tunnel_dash()` - Builds the tunnel-dash binary if needed

### `interactive.sh`
- `interactive_connection()` - Interactive prompt for connection setup

## Usage

The main `run` script sources all modules from this directory. Each module is self-contained and can be modified independently.

