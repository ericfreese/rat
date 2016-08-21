package rat

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

type configurer struct {
}

func NewConfigurer() Configurer {
	c := &configurer{}

	return c
}

func (c *configurer) ParseLine(line string) (string, []string) {
	redundantWhitespace := regexp.MustCompile(`[\s\p{Zs}]{2,}`)

	pieces := strings.SplitN(redundantWhitespace.ReplaceAllString(line, " "), " -- ", 2)
	parts := strings.Split(pieces[0], " ")

	directive := parts[0]
	args := append(parts[1:], pieces[1:]...)

	return directive, args
}

func (c *configurer) Process(rd io.Reader) {
	scanner := bufio.NewScanner(rd)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}

		directive, args := c.ParseLine(line)

		switch directive {
		case "bindkey":
			c.ProcessBindkey(args)
		case "mode":
			c.ProcessMode(scanner, args)
		default:
			panic(fmt.Sprintf("Unknown directive: '%s'", directive))
		}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
}

func (c *configurer) ProcessBindkey(args []string) {
	if len(args) != 3 {
		panic("Expected 3 args for 'bindkey'")
	}
	AddEventListener(args[0], func() {
		PushPager(NewCmdPager(args[1], args[2], Context{}))
	})
}

func (c *configurer) ProcessMode(scanner *bufio.Scanner, args []string) {
	if len(args) != 1 {
		panic("Expected 1 arg for 'mode'")
	}

	mode := NewMode()

loop:
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}

		directive, args := c.ParseLine(line)

		switch directive {
		case "annotate":
			c.ProcessModeAnnotate(mode, args)
		case "bindkey":
			c.ProcessModeBindkey(mode, args)
		case "end":
			break loop
		default:
			panic(fmt.Sprintf("Unknown directive: '%s'", directive))
		}
	}

	RegisterMode(args[0], mode)
}

func (c *configurer) ProcessModeAnnotate(mode Mode, args []string) {
	if len(args) != 3 {
		panic("Expected 3 args for 'annotate'")
	}

	if args[0] != "match" {
		panic("Annotation type must be 'match'")
	}

	mode.RegisterAnnotator(func(ctx Context) Annotator {
		return NewMatchParser(
			InterpolateContext(args[2], ctx),
			args[1],
		)
	})
}

func (c *configurer) ProcessModeBindkey(mode Mode, args []string) {
	switch len(args) {
	case 2:
		if strings.HasPrefix(args[1], "?!") {
			args[1] = args[1][2:]
			c.ProcessModeBindkeyConfirmExec(mode, args)
		} else if strings.HasPrefix(args[1], "!") {
			args[1] = args[1][1:]
			c.ProcessModeBindkeyExec(mode, args)
		}
	case 3:
		if strings.HasPrefix(args[2], "?!") {
			args[2] = args[2][2:]
			c.ProcessModeBindkeyAnnotationConfirmExec(mode, args)
		} else if strings.HasPrefix(args[2], "!") {
			args[2] = args[2][1:]
			c.ProcessModeBindkeyAnnotationExec(mode, args)
		}
	case 4:
		if strings.HasPrefix(args[3], ">") {
			args[3] = args[3][1:]
			c.ProcessModeBindkeyAnnotationChildPager(mode, args)
		} else {
			c.ProcessModeBindkeyAnnotationPushPager(mode, args)
		}
	default:
		panic("Expected 2-4 args for 'bindkey'")
	}
}

func (c *configurer) ProcessModeBindkeyConfirmExec(mode Mode, args []string) {
	mode.RegisterEventListener(func(ctx Context) func(Pager) {
		return func(p Pager) {
			p.AddEventListener(args[0], func() {
				ConfirmExec(args[1], ctx, func() {
					p.Reload()
				})
			})
		}
	})
}

func (c *configurer) ProcessModeBindkeyExec(mode Mode, args []string) {
	mode.RegisterEventListener(func(ctx Context) func(Pager) {
		return func(p Pager) {
			p.AddEventListener(args[0], func() {
				Exec(args[1], ctx)
				p.Reload()
			})
		}
	})
}

func (c *configurer) ProcessModeBindkeyAnnotationConfirmExec(mode Mode, args []string) {
	mode.RegisterEventListener(func(ctx Context) func(Pager) {
		return func(p Pager) {
			p.AddAnnotationEventListener(args[0], args[1:2], func(ectx Context) {
				ConfirmExec(args[2], MergeContext(ctx, ectx), func() {
					p.Reload()
				})
			})
		}
	})
}

func (c *configurer) ProcessModeBindkeyAnnotationExec(mode Mode, args []string) {
	mode.RegisterEventListener(func(ctx Context) func(Pager) {
		return func(p Pager) {
			p.AddAnnotationEventListener(args[0], args[1:2], func(ectx Context) {
				Exec(args[2], MergeContext(ctx, ectx))
				p.Reload()
			})
		}
	})
}

func (c *configurer) ProcessModeBindkeyAnnotationChildPager(mode Mode, args []string) {
	mode.RegisterEventListener(func(ctx Context) func(Pager) {
		return func(p Pager) {
			p.AddAnnotationEventListener(args[0], args[1:2], func(ectx Context) {
				child := NewCmdPager(
					args[2],
					args[3],
					MergeContext(ctx, ectx),
				)

				AddChildPager(p, child, args[0])
			})
		}
	})
}

func (c *configurer) ProcessModeBindkeyAnnotationPushPager(mode Mode, args []string) {
	mode.RegisterEventListener(func(ctx Context) func(Pager) {
		return func(p Pager) {
			p.AddAnnotationEventListener(args[0], args[1:2], func(ectx Context) {
				PushPager(NewCmdPager(
					args[2],
					args[3],
					MergeContext(ctx, ectx),
				))
			})
		}
	})
}
