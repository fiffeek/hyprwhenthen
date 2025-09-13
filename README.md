<img width="90" style="margin-right:10px" align=left alt="hyprwhenthen logo" src="https://github.com/user-attachments/assets/34b8648e-ee64-4f33-97d1-dbfa19cd8b01" />
<H1>HyprWhenThen</H1><br>

Event-driven automation for Hyprland. HyprWhenThen listens to Hyprland events and executes actions based on configurable rules, enabling dynamic window management and workspace automation.

## Demo

https://github.com/user-attachments/assets/528588b2-0cd1-40c3-bb21-14a67c58da5d

This demo showcases `HyprWhenThen`'s core capabilities:
- **React to any Hyprland event**, silly examples:
  - Real-time notifications when applications change fullscreen status
  - Window tracking with instant alerts on window opening
- **Live configuration reloading** without service restart

## Documentation

<!--ts-->
* [HyprWhenThen](#hyprwhenthen)
   * [Demo](#demo)
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
         * [Processing all events serially](#processing-all-events-serially)
      * [Validate](#validate)
   * [Architecture](#architecture)
      * [Routing and Concurrency](#routing-and-concurrency)
   * [Development](#development)
      * [Prerequisites](#prerequisites)
      * [Building](#building)
   * [Requirements](#requirements)
   * [License](#license)
   * [Related Projects](#related-projects)
      * [Event Automation Tools](#event-automation-tools)
      * [Advantages of hyprwhenthen](#advantages-of-hyprwhenthen)
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
then = "hyprctl dispatch togglefloating address:0x${REGEX_GROUP_1}"

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
hot_reload_debounce_timer = "100ms"  # Debounce time for config reloading, defaults to 1s
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

HyprWhenThen supports **any** Hyprland event without requiring specific parsing logic. Events follow the format `${TYPE}>>${CONTEXT}` as defined in the [Hyprland IPC specification](https://wiki.hypr.land/IPC/).

- `handler.on` matches against `${TYPE}` (the event name)
- `handler.when` regex pattern matches against `${CONTEXT}` (the event data)

**Common Event Types:**
- `windowtitlev2` - Window title changes
- `openwindow` - New window opens
- `closewindow` - Window closes
- `workspace` - Workspace changes
- `focusedmon` - Monitor focus changes
- `activewindow` - Active window changes
- `fullscreen` - Fullscreen state changes
- `monitorremoved` / `monitoradded` - Monitor connection changes
- `createworkspace` / `destroyworkspace` - Workspace lifecycle
- `moveworkspace` - Workspace moves between monitors

**Examples:**

```toml
# Window management
[[handler]]
on = "openwindow"
when = "firefox"
then = "hyprctl dispatch workspace 2"

# Monitor events
[[handler]]
on = "monitoradded"
when = "(.*)"
then = "notify-send \"Monitor connected: $REGEX_GROUP_1\""

# Custom events (future-proof)
[[handler]]
on = "monitordisabled"  # hypothetical future event
when = "(.*),(.*)"      # capture name and description
then = "notify-send \"Monitor disabled: $REGEX_GROUP_1 ($REGEX_GROUP_2)\""
```

For a complete list of current events, see the [Hyprland IPC documentation](https://wiki.hypr.land/IPC/).

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
then = "hyprctl dispatch focuswindow address:0x${REGEX_GROUP_1}"
```

### Notifications

```toml
# Notify when specific apps open
[[handler]]
on = "openwindow"
when = "(.*)"
then = "notify-send 'App opened' \"$REGEX_GROUP_1\""
```

### Serial Processing

```toml
# Process all window title changes for the same window in order
[[handler]]
on = "windowtitlev2"
when = "(.*),.*"
then = "echo \"Processing window $REGEX_GROUP_1\" >> /tmp/window.log"
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
      --queue int     Events are queued for each worker, this defines the queue size; the dispatcher will wait for a free slot when the worker is running behind (default 10)
      --workers int   Number of background workers (default 2)

Global Flags:
      --config string   Path to configuration file (default "$HOME/.config/hyprwhenthen/config.toml")
      --debug           Enable debug logging
```
<!-- END runhelp -->

#### Processing all events serially

If you want to process all events serially you could either give all of them the same `routing_key` or run
the binary with `--workers 1`. The latter ensures that only `1` event is processed at any given time.

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

### Event Automation Tools
- [**Pyprland**](https://github.com/hyprland-community/pyprland) - Full-featured Python automation framework with plugins for Hyprland window management, scratchpads, and workspace management
- [**Shellevents**](https://github.com/hyprwm/contrib/tree/main/shellevents) - Lightweight bash script that invokes shell functions in response to Hyprland socket2 events
- [**hyprevents**](https://github.com/vilari-mickopf/hyprevents) - Enhanced fork of shellevents with additional features and improvements
- [**hypr-eventsockets**](https://github.com/hyprwm/contrib/tree/main/hypr-eventsockets) - Collection of event-driven scripts for Hyprland automation

### Advantages of `hyprwhenthen`

- Regex pattern matching: Flexible filtering with capture groups for dynamic actions
- Concurrent processing: Multi-worker architecture for high-performance automation
- Minimal dependencies: Single binary with no runtime dependencies
- Events processing ordering: You can define ordering dependencies
- Run any script: It is up to the user what to run, can be any script, the event context can be automatically captured
