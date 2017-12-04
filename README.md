# Rat

_Compose shell commands to build terminal applications_

[![Join the chat at https://gitter.im/rat-chat/Lobby](https://badges.gitter.im/rat-chat/Lobby.svg)](https://gitter.im/rat-chat/Lobby?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
[![GoDoc](https://godoc.org/github.com/ericfreese/rat?status.svg)](https://godoc.org/github.com/ericfreese/rat)
[![Build Status](https://travis-ci.org/ericfreese/rat.svg?branch=master)](https://travis-ci.org/ericfreese/rat)

![](demo.gif)

## Overview

Rat was developed as part of an effort to build a [tig](https://github.com/jonas/tig)-like application with very little opinionated UI logic, delegating instead to the capabilities of shell commands like `git log` with its `--pretty` and `--graph` options.

Shell commands are executed and the output is captured and displayed in pagers. Configurable annotators parse through the output, adding annotations that can be acted upon to run other shell commands.

## Getting Started

**WARNING: THIS IS ALL SUPER EXPERIMENTAL AND IS PROBABLY GOING TO CHANGE A LOT**

### Install

```shell
$ go get github.com/ericfreese/rat
$ go build && go install
```

### Configure

Rat is configured through a file `ratrc` in your home config directory ([`$XDG_CONFIG_HOME/rat`](https://specifications.freedesktop.org/basedir-spec/latest), `~/.config/rat` by default).

Rat pagers can be opened in one or more "modes". A mode is a configuration of "annotators" and "key bindings":

- Annotators will look through the content of the pager for special bits of text that actions can be taken on. These bits of text are called "annotations". Each annotation has a start, an end, a class, and a value.
- Key bindings define actions that can be taken on annotations.

#### Keybindings

First you'll need to set up some keybindings. Add the following to your `ratrc` and modify as desired:

```shell
bindkey C-r reload
bindkey j   cursor-down
bindkey k   cursor-up
bindkey C-e scroll-down
bindkey C-y scroll-up
bindkey C-d page-down
bindkey C-u page-up
bindkey g,g cursor-first-line
bindkey S-g cursor-last-line
bindkey S-j parent-cursor-down
bindkey S-k parent-cursor-up
bindkey q   pop-pager
bindkey S-q quit
bindkey M-1 show-one
bindkey M-2 show-two
bindkey M-3 show-three
```

<kbd>ctrl</kbd>+<kbd>c</kbd> will always quit.

#### Mode Definitions

The `mode` keyword starts a mode definition.

```shell
mode <name>
  ...
end
```

#### Annotator Definitions

Inside of a mode definition, the `annotate` keyword starts an annotation definition.

```shell
mode <name>
  annotate <type> <class> -- <options>
end
```

- `type`: The annotator type. Can be "match", "regex", or "external".
- `class`: The class to apply to any annotations that this annotator finds.
- `options`:
    - If `type` is "match", this should be a shell command that outputs newline-delimited strings that the annotator will search for.
    - If `type` is "regex", this should define a regular expression to search for (Golang regular expressions are supported).
    - If `type` is "external", this should be the name of an executable located in `~/.config/rat/annotators/` that will be executed and sent the content of the pager via STDIN. The executable should print annotations to STDOUT in a specific binary format:
        - Start: Byte offset from the beginning of STDIN (64-bit little-endian unsigned integer)
        - End: Byte offset from the beginning of STDIN (64-bit little-endian unsigned integer)
        - Value length: Length of found value string in bytes (64-bit little-endian unsigned integer)
        - Value string: String of above-specified length (UTF-8 encoded)

#### Keybinding Definitions

The `bindkey` keyword starts a keybinding definition.

```shell
mode <name>
  bindkey <key> [<annotation-classes>] [<new-pager-mode>] -- <action>
end

bindkey <key> <action>
bindkey <key> <new-pager-mode> -- <cmd>
```

- `key`: A key combination that will trigger this action when pressed. Modifiers are added with `C-` and `S-`. See `lib/key_event.go` for a list of supported named keys.
- `action`: A named action to run when the key is pressed. See action.go for a list of available actions.
- `annotation-classes`: This action will only be triggered if annotations of these classes are present on the current line. If omitted, keybinding will work anywhere in the pager. These should be comma-delimited.
- `new-pager-mode`: If the action will create a new pager, this defines the mode(s) to use when creating that pager.
- `cmd`: A shell command to run when the specified key combination is pressed. Annotation values will be exported to the command process as variables named for their annotation class. The default is to open a new pager showing the output of the shell command, but several special prefixes can be used to specify different actions to be taken:
    - `!`: Do not open a new pager. Execute the command and reload the current pager.
    - `?!`: Like `!`, but confirm with the user first (will have to press 'y' for yes or 'n' for no).
    - `>`: Like the default, open a new pager with the contents of the shell command, but also set up a parent-child relationship so that the parent cursor can be moved up and down from inside the child pager with the `ParentCursorUp` and `ParentCursorDown` commands.

Note: Keybindings that are not inside of a mode definition will always be available and do not have the special prefix behavior described above.

#### Source configurations from separate files

The `source` keyword imports configuration rules from another file.

```shell
source <file>
```

- `file`: The path (relative to the rat config directory) to a file that contains valid rat configuration instructions

#### Example

Add the following to your `ratrc` to build a simple file viewer/manager:

```shell
mode files
  # Find all files (not directories) in the current directory and
  # annotate with the class "file".
  annotate match file -- ls -a1p | grep -v /

  # When the cursor is on a line with an annotation of class "file" and
  # the `enter` key is pressed, run `cat` with the value of the
  # annotation (the filename) and display the output in a new pager with
  # mode "preview". 
  bindkey enter file preview -- >cat $file

  # When the cursor is on a line with an annotation of class "file" and
  # the `e` key is pressed, open the selected file in vim.
  bindkey e     file         -- !vim $file

  # When the cursor is on a line with an annotation of class "file" and
  # Shift + `x` is pressed, delete the file if the user confirms it.
  bindkey S-x   file         -- ?!rm $file
end
```

Run `rat --mode files --cmd 'ls -al'` to try it out. You should see the output of `ls -al`. Move your cursor to a line with a regular file on it and press <kbd>Enter</kbd> to view its contents. Try out the other keybindings. Try tweaking some things.

See `examples/` directory for more configuration examples.

## Usage

### Run

``` shell
rat [--mode=<mode>] [--cmd=<command>]
```

`--mode` defaults to `default`.

If `--cmd` is not provided, rat will read from STDIN.

## Development

### Dependencies

Dependencies are managed using [`glide`](https://github.com/Masterminds/glide).

## License

This project is licensed under [MIT license](http://opensource.org/licenses/MIT). For the full text of the license, see the [LICENSE](LICENSE) file.
