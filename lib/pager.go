package rat

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type Pager interface {
	Widget
	Window
	AddAnnotator(Annotator)
	AddEventHandler(string, EventHandler)
	Reload()
	GetContext() Context
	GetWorkingDir() string
	AddDestroyHook(func())
	MoveCursorNext(string)
	MoveCursorPrevious(string)
}

type pager struct {
	title         string
	modes         []Mode
	ctx           Context
	buffer        Buffer
	annotators    []Annotator
	eventHandlers HandlerRegistry
	destroyHooks  []func()
	*pagerLayout
	Window
}

func newPager(title, modeNames string, ctx Context) *pager {
	p := &pager{}

	p.title = title
	p.ctx = ctx
	p.eventHandlers = NewHandlerRegistry()
	p.pagerLayout = &pagerLayout{}
	p.annotators = make([]Annotator, 0, 8)
	p.destroyHooks = make([]func(), 0, 0)

	p.Window = NewWindow(
		func() int { return p.GetContentBox().Height() },
		func() int { return p.buffer.NumLines() },
	)

	splitModeNames := strings.Split(modeNames, ",")
	p.modes = make([]Mode, 0, len(splitModeNames))

	for _, modeName := range splitModeNames {
		if mode, ok := modes[modeName]; ok {
			p.modes = append(p.modes, mode)
		}
	}

	return p
}

func (p *pager) AddAnnotator(a Annotator) {
	p.annotators = append(p.annotators, a)

	if p.buffer != nil {
		go p.buffer.AnnotateWith(a)
	}
}

func (p *pager) AddEventHandler(keyStr string, handler EventHandler) {
	p.eventHandlers.Add(KeySequenceFromString(keyStr), handler)
}

func (p *pager) AddDestroyHook(hook func()) {
	p.destroyHooks = append(p.destroyHooks, hook)
}

func (p *pager) Stop() {
	p.buffer.Close()
}

func (p *pager) Destroy() {
	for _, hook := range p.destroyHooks {
		hook()
	}

	p.Stop()
}

func (p *pager) HandleEvent(ks []keyEvent) bool {
	p.buffer.Lock()
	defer p.buffer.Unlock()

	ctx := NewContextFromAnnotations(p.buffer.AnnotationsForLine(p.Window.GetCursor()))

	if handler := p.eventHandlers.FindCtx(ks, ctx); handler != nil {
		handler.Call(ctx)
		return true
	}

	return false
}

func (p *pager) Render() {
	p.buffer.Lock()
	defer p.buffer.Unlock()

	p.drawHeader(
		p.title,
		fmt.Sprintf("%d %d/%d", p.buffer.NumAnnotations(), p.Window.GetCursor()+1, p.buffer.NumLines()),
	)

	p.drawContent(
		p.Window.GetCursor()-p.Window.GetScroll(),
		p.buffer.StyledLines(p.Window.GetScroll(), p.GetContentBox().Height()),
	)
}

func (p *pager) Reload() {
}

func (p *pager) GetContext() Context {
	return p.ctx
}

func (p *pager) GetWorkingDir() string {
	wd, _ := os.Getwd()
	return wd
}

func (p *pager) MoveCursorNext(annotationClass string) {
	next := p.buffer.FindNextAnnotation(p.GetCursor(), annotationClass)

	if next >= 0 {
		p.MoveCursorTo(next)
	}
}

func (p *pager) MoveCursorPrevious(annotationClass string) {
	p.MoveCursorTo(p.buffer.FindPreviousAnnotation(p.GetCursor(), annotationClass))
}

func NewReadPager(rd io.Reader, title string, modeNames string, ctx Context) Pager {
	p := newPager(title, modeNames, ctx)

	for _, mode := range p.modes {
		mode.DecoratePager(p)
	}

	p.buffer = NewBuffer(rd, p.annotators)

	return p
}

type cmdPager struct {
	*pager
	cmd     string
	command ShellCommand
	wd      string
}

func NewCmdPager(modeNames, cmd, wd string, ctx Context) Pager {
	cp := &cmdPager{}

	cp.cmd = cmd
	cp.wd = wd
	cp.pager = newPager(cmd, modeNames, ctx)

	for _, mode := range cp.modes {
		mode.DecoratePager(cp)
	}

	cp.RunCommand()

	return cp
}

func (cp *cmdPager) Stop() {
	cp.command.Close()
	cp.pager.Stop()
}

func (cp *cmdPager) Reload() {
	cp.Stop()
	cp.RunCommand()
}

func (cp *cmdPager) RunCommand() {
	var err error

	if cp.command, err = NewShellCommand(cp.cmd, cp.wd, cp.ctx); err != nil {
		panic(err)
	}

	cp.buffer = NewBuffer(cp.command, cp.pager.annotators)
}

func (cp *cmdPager) GetWorkingDir() string {
	return cp.wd
}
