# Zellijer
Frontend to a start Zellij by either attaching to existing sessions or using a layout.

Zellijer is intended to be a program that could be launched when creating a toplevel terminal. It provides a choice of zellij sessions to attach to and a list of layouts to start a new session. 

Basically, I saw a pain point when I started using a multiplexer and my first thought was to create rofi/dmenu interface. But I have been wanting to try [Bubbletea](https://github.com/charmbracelet/bubbletea/) for a while and this seemed like a good fit. Bubbles List is basically dmenu in TUI form already.

## Features

- Layout list is currently only top level in one zellij layout directory
- List current sessions

## Usage

### Fish

Add this to your fish config file

```fish
if set -q ZELLIJ
else
    zellijer
end
```

## TODO

- Update in app help/hints
- Add keys for common use cases
- Use spinner for loading
- Name sessions before creation
- Get layouts recursively
- Get layouts from multiple directories
- Support manually adding layouts

## Possible future

- Support other multiplexers
- Edit layouts
- Generate layouts from a session

