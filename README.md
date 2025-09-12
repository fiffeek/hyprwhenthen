# HyprWhenThen

Event-driven automation for Hyprland. HyprWhenThen listens to Hyprland events and executes actions based on configurable rules, enabling dynamic window management and workspace automation.

## Documentation

<!--ts-->
* [HyprWhenThen](#hyprwhenthen)
   * [Documentation](#documentation)
   * [Features](#features)
   * [Installation](#installation)
      * [Binary Release](#binary-release)
      * [AUR](#aur)
      * [Build from Source](#build-from-source)
   * [Quick start](#quick-start)
      * [Basic Configuration](#basic-configuration)
      * [Run service](#run-service)
   * [Configuration](#configuration)
      * [General Section](#general-section)
      * [Handlers](#handlers)
         * [Supported Events](#supported-events)
         * [Template Variables](#template-variables)
         * [Routing Keys](#routing-keys)
   * [Examples](#examples)
      * [Window Management](#window-management)
      * [Dynamic Workspace Switching](#dynamic-workspace-switching)
      * [Notifications](#notifications)
      * [Serial Processing](#serial-processing)
   * [Command Line Options](#command-line-options)
      * [Run](#run)
      * [Validate](#validate)
   * [Architecture](#architecture)
      * [Routing and Concurrency](#routing-and-concurrency)
   * [Development](#development)
      * [Prerequisites](#prerequisites)
      * [Building](#building)
   * [Requirements](#requirements)
   * [License](#license)
   * [Related Projects](#related-projects)
<!--te-->

## Features

- Event-driven automation: React to any Hyprland event (window title changes, workspace switches, etc.)
- Regex pattern matching: Flexible event filtering with capture groups for dynamic actions
- Concurrent execution: Multi-worker architecture with configurable parallelism
- Routing keys: Control execution order for related events
- Hot configuration reloading: Update rules without restarting the service
- Timeout management: Configurable timeouts for both global and per-handler execution
- Template variables: Use regex capture groups in your action commands

## Installation

### Binary Release

Download the latest binary from GitHub releases:

```bash
# optionally override the destination directory, defaults to ~/.local/bin/
export DESTDIR="$HOME/.bin"
curl -o- https://raw.githubusercontent.com/fiffeek/hyprwhenthen/refs/heads/main/scripts/install.sh | bash
```

### AUR

For Arch Linux users, install from the AUR:

```bash
# Using your preferred AUR helper (replace 'aurHelper' with your choice)
aurHelper="yay"  # or paru, trizen, etc.
$aurHelper -S hyprwhenthen-bin

# Or using makepkg:
git clone https://aur.archlinux.org/hyprwhenthen-bin.git
cd hyprwhenthen-bin
makepkg -si
```

### Build from Source

Requires [asdf](https://asdf-vm.com/) to manage the Go toolchain:
```bash
# Build the binary (output goes to ./dest/)
make

# Install to custom location
make DESTDIR=$HOME/binaries install

# Uninstall from custom location
make DESTDIR=$HOME/binaries uninstall

# Install system-wide (may require sudo)
sudo make DESTDIR=/usr/bin install
````

## Quick start

### Basic Configuration

Create `~/.config/hyprwhenthen/config.toml`:

```toml
[general]
timeout = "1s"

# Float Google login window automatically
[[handler]]
on = "windowtitlev2"
when = "(.*),Sign In - Google Account"
then = "hyprctl dispatch togglefloating address:$REGEX_GROUP_1"

# Switch to workspace when specific app opens
[[handler]]
on = "openwindow"
when = "firefox"
then = "hyprctl dispatch workspace 2"
```

### Run service

```bash
# Start the service
hyprwhenthen run

# Validate configuration
hyprwhenthen validate

# Run with debug logging
hyprwhenthen run --debug
```

## Configuration

### General Section

```toml
[general]
timeout = "15s"                      # Global timeout for all handlers
hot_reload_debounce_timer = "100ms"  # Debounce time for config reloading
```

### Handlers

Each handler defines an event-action rule:

```toml
[[handler]]
on = "windowtitlev2"                 # Hyprland event type
when = "(.*),Mozilla Firefox"       # Regex pattern to match event data
then = "notify-send 'Firefox: $REGEX_GROUP_1'"  # Command to execute
timeout = "5s"                       # Optional: override global timeout
routing_key = "$REGEX_GROUP_1"       # Optional: control execution order
```

#### Supported Events

HyprWhenThen supports any hyprland event, as it does not parse them in any capacity.
The only thing required is that the events are of format: `${TYPE}>>${CONTEXT}` (which is
with accordance to the [spec](https://wiki.hypr.land/IPC/). Anything that the user defines in
`handler.on` will be matched to `${TYPE}`, and `handler.when` regex is matched with `${CONTEXT}`.

For instance, if hyprland introduces a new event: `monitordisabled>>NAME,DESCRIPTION`,
you could capture it with:
```toml
[[handler]]
on = "monitordisabled"
when = "(.*),(.*)" # matches both groups so that any script can use them
then = "notify-send \"Monitor disabled, name: $REGEX_GROUP_1, desc: $REGEX_GROUP_2\""
```

#### Template Variables

Use regex capture groups in your commands:

- `$REGEX_GROUP_0` - Full matched string
- `$REGEX_GROUP_1` - First capture group
- `$REGEX_GROUP_2` - Second capture group
- etc.

The environment for the commands is the same as the one that the service is running in,
plus all the above variables from pattern matching.

#### Routing Keys

Control execution order for related events by using routing keys. Events with the same routing key are processed serially:

```toml
# All events for the same window address are processed in order
[[handler]]
on = "windowtitlev2"
when = "(.*),.*"
then = "echo 'Window title changed: $REGEX_GROUP_1'"
routing_key = "$REGEX_GROUP_1"  # Window address
```

You can use any known environment variables in the `routing_key` or a plain string.
Omitting the `routing_key` results in random worker allocation.

## Examples

Some of these can be achieved with pure hyprland configuration.
If you want to see a more comprehensive example that *can't* be achieved
purely by hypr config, see [how to float a window that changes title runtime](https://github.com/fiffeek/hyprwhenthen/blob/main/examples/minimal/config.toml).

### Window Management

```toml
# Auto-float specific windows
[[handler]]
on = "openwindow"
when = "pavucontrol|calculator"
then = "hyprctl dispatch togglefloating"

# Move Discord to workspace 9
[[handler]]
on = "openwindow"
when = "discord"
then = "hyprctl dispatch movetoworkspacesilent 9"
```

### Dynamic Workspace Switching

```toml
# Follow Firefox windows to their workspace
[[handler]]
on = "windowtitlev2"
when = "(.*),.*Mozilla Firefox"
then = "hyprctl dispatch focuswindow address:$REGEX_GROUP_1"
```

### Notifications

```toml
# Notify when specific apps open
[[handler]]
on = "openwindow"
when = "(.*)"
then = "notify-send 'App opened' '$REGEX_GROUP_1'"
```

### Serial Processing

```toml
# Process all window title changes for the same window in order
[[handler]]
on = "windowtitlev2"
when = "(.*),.*"
then = "echo 'Processing window $REGEX_GROUP_1' >> /tmp/window.log"
routing_key = "$REGEX_GROUP_1"
```

## Command Line Options
<!-- START help -->
```text
HyprWhenThen is an automation tool that listens to Hyprland events and executes actions based on configured rules.

Usage:
  hyprwhenthen [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  run         Start the HyprWhenThen service
  validate    Validate configuration file

Flags:
      --config string   Path to configuration file (default "$HOME/.config/hyprwhenthen/config.toml")
      --debug           Enable debug logging
  -h, --help            help for hyprwhenthen
  -v, --version         version for hyprwhenthen

Use "hyprwhenthen [command] --help" for more information about a command.
```
<!-- END help -->

### Run
<!-- START runhelp -->
```text
Start the HyprWhenThen service to listen for Hyprland events and execute configured actions.

Usage:
  hyprwhenthen run [flags]

Flags:
  -h, --help          help for run
      --queue int     Defines the queue size (default 10)
      --workers int   Number of background workers (default 2)

Global Flags:
      --config string   Path to configuration file (default "$HOME/.config/hyprwhenthen/config.toml")
      --debug           Enable debug logging
```
<!-- END runhelp -->

### Validate
<!-- START validatehelp -->
```text
Validate the syntax and structure of the HyprWhenThen configuration file.

Usage:
  hyprwhenthen validate [flags]

Flags:
  -h, --help   help for validate

Global Flags:
      --config string   Path to configuration file (default "$HOME/.config/hyprwhenthen/config.toml")
      --debug           Enable debug logging
```
<!-- END validatehelp -->


## Architecture

HyprWhenThen uses a multi-worker architecture for concurrent event processing.

### Routing and Concurrency

- Events without routing keys are distributed randomly across workers
- Events with the same routing key are processed serially by the same worker

## Development

### Prerequisites

- asdf (for development environment)

### Building

```bash
# Set up development environment
make dev

# Run tests
make test/integration

# Build binary
make build/test

# Lint and format
make lint
make fmt
```


## Requirements

- **Hyprland**: Version with IPC support

## License

MIT License - see [LICENSE](LICENSE) for details.

## Related Projects

- [Pyprland](https://github.com/hyprland-community/pyprland) - Pure hypr IPC Python automation
- [Shellevents](https://github.com/hyprwm/contrib/tree/main/shellevents) - Invoke shell functions in response to Hyprland socket2 events
- [hyprevents](https://github.com/vilari-mickopf/hyprevents/tree/master) - Fork of the above
