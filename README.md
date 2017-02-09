# Rat

_Compose shell commands to build terminal applications_

![](demo.gif)

## Overview

Rat was developed as part of an effort to build a [tig](https://github.com/jonas/tig)-like application with very little opinionated UI logic, delegating instead to the capabilities of shell commands like `git log` with its `--pretty` and `--graph` options.

Shell commands are executed and the output is captured. Configurable annotators parse through the output, adding annotations that can then be acted upon by the user to run other shell commands.

## Getting Started

### Install

```shell
$ go get github.com/ericfreese/rat
$ go build && go install
```

### Configure

Create a file `.ratrc` in your home directory and define modes (annotators + key bindings) and top-level key bindings.

Simple example:

```shell
mode helloworld
  annotate match lower -- echo "hello\nworld"
  annotate match upper -- echo "HELLO\nWORLD"

  bindkey enter lower helloworld -- >echo "%(lower)" | awk '{print toupper($0)}'
  bindkey enter upper helloworld -- >echo "%(upper)" | awk '{print tolower($0)}'
end
```

Run `rat --mode helloworld --cmd 'echo "hello\nworld\nHELLO\nWORLD"'` to try it out.

See `examples/` directory for more configuration examples.

## Usage

### Run

``` shell
rat [--mode=<mode>] [--cmd=<command>]
```

`--mode` defaults to `default`
`--cmd` defaults to `cat ~/.ratrc`



### Keybindings

```golang
// Pager key bindings
p.AddEventListener("C-r", p.Reload)
p.AddEventListener("j", p.CursorDown)
p.AddEventListener("k", p.CursorUp)
p.AddEventListener("down", p.CursorDown)
p.AddEventListener("up", p.CursorUp)
p.AddEventListener("C-j", p.ScrollDown)
p.AddEventListener("C-k", p.ScrollUp)
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

## License

This project is licensed under [MIT license](http://opensource.org/licenses/MIT). For the full text of the license, see the [LICENSE](LICENSE) file.
