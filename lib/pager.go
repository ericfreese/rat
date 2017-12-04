package rat

import (
	"fmt"
	"io"
	"strings"

	termbox "github.com/nsf/termbox-go"
)

type Pager interface {
	Widget
	AddEventHandler(keyStr string, handler EventHandler)
	Reload()
	CursorUp()
	CursorDown()
	CursorFirstLine()
	CursorLastLine()
	ScrollUp()
	ScrollDown()
	PageUp()
	PageDown()
}

type pager struct {
	title         string
	modes         []Mode
	ctx           Context
	buffer        Buffer
	scrollOffsetY int
	cursorY       int
	stop          chan bool
	eventHandlers HandlerRegistry

	box        Box
	headerBox  Box
	contentBox Box
}

func newPager(title string, modeNames string, ctx Context) *pager {
	p := &pager{}

	p.title = title
	p.ctx = ctx
	p.eventHandlers = NewHandlerRegistry()

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

	ctx := NewContextFromAnnotations(p.buffer.AnnotationsForLine(p.cursorY))

	if handler := p.eventHandlers.FindCtx(ks, ctx); handler != nil {
		handler.Call(ctx)
		return true
	}

	return false
}

func (p *pager) SetBox(box Box) {
	p.box = box
	p.layout()
}

func (p *pager) GetBox() Box {
	return p.box
}

func (p *pager) layout() {
	p.headerBox = NewBox(p.box.Left(), p.box.Top(), p.box.Width(), 1)
	p.contentBox = NewBox(p.box.Left(), p.box.Top()+1, p.box.Width(), p.box.Height()-1)
}

func (p *pager) drawHeader() {
	p.headerBox.DrawStyledRunes(1, 0, StyledRunesFromString(p.title, gTermStyles.Get(termbox.AttrUnderline, termbox.ColorDefault)))

	pagerInfo := StyledRunesFromString(fmt.Sprintf(" %d %d/%d ", p.buffer.NumAnnotations(), p.cursorY+1, p.buffer.NumLines()), gTermStyles.Get(termbox.AttrBold, termbox.ColorDefault))
	p.headerBox.DrawStyledRunes(p.headerBox.Width()-len(pagerInfo), 0, pagerInfo)
}

func (p *pager) drawContent() {
	p.contentBox.DrawStyledRune(1, p.cursorY-p.scrollOffsetY, NewStyledRune('▶', gTermStyles.Get(termbox.ColorRed, termbox.ColorDefault)))

	for y, line := range p.buffer.StyledLines(p.scrollOffsetY, p.contentBox.Height()) {
		p.contentBox.DrawStyledRunes(3, y, []StyledRune(line))
	}
}

func (p *pager) Render() {
	p.buffer.Lock()
	p.drawHeader()
	p.drawContent()
	p.buffer.Unlock()
}

func (p *pager) MoveCursorToY(cursorY int) {
	if cursorY < 0 {
		p.cursorY = 0
	} else if cursorY >= p.buffer.NumLines() {
		p.cursorY = p.buffer.NumLines() - 1
	} else {
		p.cursorY = cursorY
	}

	if p.cursorY < p.scrollOffsetY {
		p.ScrollToY(p.cursorY)
	} else if p.cursorY > p.scrollOffsetY+p.contentBox.Height()-1 {
		p.ScrollToY(p.cursorY - (p.contentBox.Height() - 1))
	}
}

func (p *pager) MoveCursorY(delta int) {
	p.MoveCursorToY(p.cursorY + delta)
}

func (p *pager) ScrollToY(scrollY int) {
	if scrollY < 0 {
		p.scrollOffsetY = 0
	} else if scrollY >= p.buffer.NumLines()-p.contentBox.Height() {
		if p.buffer.NumLines() > p.contentBox.Height() {
			p.scrollOffsetY = p.buffer.NumLines() - p.contentBox.Height()
		} else {
			p.scrollOffsetY = 0
		}
	} else {
		p.scrollOffsetY = scrollY
	}

	if p.cursorY < p.scrollOffsetY {
		p.MoveCursorToY(p.scrollOffsetY)
	} else if p.cursorY > p.scrollOffsetY+p.contentBox.Height()-1 {
		p.MoveCursorToY(p.scrollOffsetY + p.contentBox.Height() - 1)
	}
}

func (p *pager) ScrollY(delta int) {
	p.ScrollToY(p.scrollOffsetY + delta)
}

func (p *pager) CursorY() int {
	return p.cursorY
}

func (p *pager) CursorUp() {
	p.MoveCursorY(-1)
}

func (p *pager) CursorDown() {
	p.MoveCursorY(1)
}

func (p *pager) CursorFirstLine() {
	p.MoveCursorToY(0)
}

func (p *pager) CursorLastLine() {
	p.MoveCursorToY(p.buffer.NumLines())
}

func (p *pager) ScrollUp() {
	p.ScrollY(-1)
}

func (p *pager) ScrollDown() {
	p.ScrollY(1)
}

func (p *pager) PageUp() {
	p.ScrollY(-p.contentBox.Height())
}

func (p *pager) PageDown() {
	p.ScrollY(p.contentBox.Height())
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
