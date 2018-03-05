package rat

import (
	"fmt"
	"io"
	"strings"

	termbox "github.com/nsf/termbox-go"
)

type PagerLayout interface {
	Layout
	GetHeaderBox() Box
	GetContentBox() Box
}

type pagerLayout struct {
	box        Box
	headerBox  Box
	contentBox Box
}

func (pl *pagerLayout) layout() {
	pl.headerBox = NewBox(pl.box.Left(), pl.box.Top(), pl.box.Width(), 1)
	pl.contentBox = NewBox(pl.box.Left(), pl.box.Top()+1, pl.box.Width(), pl.box.Height()-1)
}

func (pl *pagerLayout) SetBox(box Box) {
	pl.box = box
	pl.layout()
}

func (pl *pagerLayout) GetBox() Box {
	return pl.box
}

func (pl *pagerLayout) GetHeaderBox() Box {
	return pl.headerBox
}

func (pl *pagerLayout) GetContentBox() Box {
	return pl.contentBox
}

func (pl *pagerLayout) drawHeader(title, info string) {
	paddedInfo := StyledRunesFromString(fmt.Sprintf(" %s ", info), gTermStyles.Get(termbox.AttrBold, termbox.ColorDefault))

	pl.GetHeaderBox().DrawStyledRunes(1, 0, StyledRunesFromString(title, gTermStyles.Get(termbox.AttrUnderline, termbox.ColorDefault)))
	pl.GetHeaderBox().DrawStyledRunes(pl.GetHeaderBox().Width()-len(paddedInfo), 0, paddedInfo)
}

func (pl *pagerLayout) drawContent(cursor int, lines [][]StyledRune) {
	pl.GetContentBox().DrawStyledRune(1, cursor, NewStyledRune('▶', gTermStyles.Get(termbox.ColorRed, termbox.ColorDefault)))

	for y, line := range lines {
		pl.contentBox.DrawStyledRunes(3, y, line)
	}
}

type Pager interface {
	Widget
	Window
	AddEventHandler(keyStr string, handler EventHandler)
	Reload()
}

type pager struct {
	title         string
	modes         []Mode
	ctx           Context
	buffer        Buffer
	stop          chan bool
	eventHandlers HandlerRegistry
	*pagerLayout
	Window
}

func newPager(title string, modeNames string, ctx Context) *pager {
	p := &pager{}

	p.title = title
	p.ctx = ctx
	p.eventHandlers = NewHandlerRegistry()
	p.pagerLayout = &pagerLayout{}

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

func (p *pager) startAnnotators() {
	for _, m := range p.modes {
		for _, a := range m.InitAnnotators(p.ctx)() {
			go p.buffer.AnnotateWith(a)
		}
	}
}

func (p *pager) AddEventHandler(keyStr string, handler EventHandler) {
	p.eventHandlers.Add(KeySequenceFromString(keyStr), handler)
}

func (p *pager) Stop() {
	p.buffer.Close()
}

func (p *pager) Destroy() {
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

func NewReadPager(rd io.Reader, title string, modeNames string, ctx Context) Pager {
	p := newPager(title, modeNames, ctx)

	for _, mode := range p.modes {
		mode.AddEventHandlers(ctx)(p)
	}

	p.buffer = NewBuffer(rd)
	p.startAnnotators()

	return p
}

type cmdPager struct {
	*pager
	cmd     string
	command ShellCommand
}

func NewCmdPager(modeNames string, cmd string, ctx Context) Pager {
	cp := &cmdPager{}

	cp.cmd = cmd
	cp.pager = newPager(cmd, modeNames, ctx)

	for _, mode := range cp.modes {
		mode.AddEventHandlers(ctx)(cp)
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

	if cp.command, err = NewShellCommand(cp.cmd, cp.ctx); err != nil {
		panic(err)
	}

	cp.buffer = NewBuffer(cp.command)
	cp.pager.startAnnotators()
}
