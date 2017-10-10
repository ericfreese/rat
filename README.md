# Rat

_Compose shell commands to build terminal applications_

[![GoDoc](https://godoc.org/github.com/ericfreese/rat?status.svg)](https://godoc.org/github.com/ericfreese/rat)

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

Rat is configured through a file `.ratrc` in your home config directory (`~/.config/rat/.ratrc`).

Rat pagers can be opened in one or more "modes". A mode is a configuration of "annotators" and "key bindings":

- Annotators will look through the content of the pager for special bits of text that actions can be taken on. These bits of text are called "annotations". Each annotation has a start, an end, a class, and a value.
- Key bindings define actions that can be taken on annotations.

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
  bindkey <key> [<annotation-class>] [<new-pager-mode>] -- <action>
end

bindkey <key> <new-pager-mode> -- <action>
```

- `key`: A key combination that will trigger this action when pressed. Modifiers are added with `C-` and `S-`. See `lib/key_event.go` for a list of supported named keys.
- `annotation-class`: This action will only be triggered on annotations of this class. If ommitted, keybinding will work anywhere in the pager.
- `new-pager-mode`: If the action will create a new pager, this defines the mode(s) to use when creating that pager.
- `action`: A shell command to run when the specified key combination is pressed. Annotation values can be interpolated into the command using `%(<annotation-class>)`. The default is to open a new pager showing the output of the shell command, but several special prefixes can be used to specify different actions to be taken:
    - `!`: Do not open a new pager. Execute the command and reload the current pager.
    - `?!`: Like `!`, but confirm with the user first (will have to press 'y' for yes or 'n' for no).
    - `>`: Like the default, open a new pager with the contents of the shell command, but also set up a parent-child relationship so that the parent cursor can be moved up and down from inside the child pager with the `ParentCursorUp` and `ParentCursorDown` commands.

Note: Keybindings that are not inside of a mode definition will always be available and do not have the special prefix behavior described above.

#### Example

Add the following to your `.ratrc` to build a simple file viewer/manager:

```shell
mode files
  # Find all files (not directories) in the current directory and
  # annotate with the class "file".
  annotate match file -- ls -a1p | grep -v /

  # When the cursor is on a line with an annotation of class "file" and
  # the `enter` key is pressed, run `cat` with the value of the
  # annotation (the filename) and display the output in a new pager with
  # mode "preview". 
  bindkey enter file preview -- >cat %(file)

  # When the cursor is on a line with an annotation of class "file" and
  # the `e` key is pressed, open the selected file in vim.
  bindkey e     file         -- !vim %(file)

  # When the cursor is on a line with an annotation of class "file" and
  # Shift + `x` is pressed, delete the file if the user confirms it.
  bindkey S-x   file         -- ?!rm %(file)
end
```

Run `rat --mode files --cmd 'ls -al'` to try it out. You should see the output of `ls -al`. Move your cursor to a line with a regular file on it and press <kbd>Enter</kbd> to view its contents. Try out the other keybindings. Try tweaking some things.

See `examples/` directory for more configuration examples.

## Usage

### Run

``` shell
rat [--mode=<mode>] [--cmd=<command>]
```

`--mode` defaults to `default`
`--cmd` defaults to `cat ~/.config/rat/.ratrc`

### Keybindings

```golang
// Pager key bindings
p.AddEventListener("C-r", p.Reload)
p.AddEventListener("j", p.CursorDown)
p.AddEventListener("k", p.CursorUp)
p.AddEventListener("down", p.CursorDown)
p.AddEventListener("up", p.CursorUp)
p.AddEventListener("C-e", p.ScrollDown)
p.AddEventListener("C-y", p.ScrollUp)
p.AddEventListener("pgdn", p.PageDown)
p.AddEventListener("pgup", p.PageUp)
p.AddEventListener("g", p.CursorFirstLine)
p.AddEventListener("S-g", p.CursorLastLine)

// Pager stack bindings
ps.AddEventListener("S-j", ps.ParentCursorDown)
ps.AddEventListener("S-k", ps.ParentCursorUp)

// Top-level key bindings
AddEventListener("q", PopPager)
AddEventListener("S-q", Quit)
AddEventListener("1", func() { pagers.Show(1) })
AddEventListener("2", func() { pagers.Show(2) })
AddEventListener("3", func() { pagers.Show(3) })
```

## Development

### Dependencies

Dependencies are managed using [`glide`](https://github.com/Masterminds/glide).

## License

This project is licensed under [MIT license](http://opensource.org/licenses/MIT). For the full text of the license, see the [LICENSE](LICENSE) file.
