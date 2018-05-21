package rat

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	termbox "github.com/nsf/termbox-go"
)

type Prompt interface {
	Widget
	Confirm(message string, callback func())
	Text(label string, callback func(text string, success bool))
}

type promptStrategy interface {
	Render(Box)
	HandleEvent([]keyEvent) bool
	Destroy()
}

type prompt struct {
	current promptStrategy
	box     Box
}

func NewPrompt() Prompt {
	return &prompt{}
}

func (p *prompt) Confirm(message string, callback func()) {
	p.current = NewConfirmPrompt(message, func(confirmed bool) {
		p.current.Destroy()
		p.current = nil

		if confirmed {
			callback()
		}
	})
}

func (p *prompt) Text(label string, callback func(string, bool)) {
	p.current = NewTextPrompt(label, func(text string, success bool) {
		p.current.Destroy()
		p.current = nil

		callback(text, success)
	})
}

func (p *prompt) SetBox(b Box) {
	p.box = b
}

func (p *prompt) GetBox() Box {
	return p.box
}

func (p *prompt) Render() {
	if p.current != nil {
		p.current.Render(p.box)
	}
}

func (p *prompt) HandleEvent(ks []keyEvent) bool {
	if p.current != nil {
		return p.current.HandleEvent(ks)
	}

	return false
}

func (p *prompt) Destroy() {
	if p.current != nil {
		p.current.Destroy()
	}
}

type confirmPrompt struct {
	message  string
	callback func(bool)
}

func NewConfirmPrompt(message string, callback func(bool)) promptStrategy {
	cp := &confirmPrompt{}

	cp.message = fmt.Sprintf("%s (y/n): ", message)
	cp.callback = callback

	return cp
}

func (cp *confirmPrompt) Render(b Box) {
	b.DrawStyledRunes(1, 0, StyledRunesFromString(cp.message, gTermStyles.Get(termbox.ColorGreen, termbox.ColorDefault)))
	termbox.SetCursor(b.Left()+1+utf8.RuneCountInString(cp.message), b.Top())
}

func (cp *confirmPrompt) HandleEvent(ks []keyEvent) bool {
	switch ks[len(ks)-1] {
	case KeyEventFromString("y"), KeyEventFromString("S-y"):
		cp.callback(true)
	case KeyEventFromString("n"), KeyEventFromString("S-n"):
		cp.callback(false)
	}

	return true
}

func (cp *confirmPrompt) Destroy() {
	termbox.HideCursor()
}

type textPrompt struct {
	label         string
	text          []rune
	callback      func(string, bool)
	origInputMode termbox.InputMode
}

func NewTextPrompt(label string, callback func(string, bool)) promptStrategy {
	tp := &textPrompt{}

	tp.label = fmt.Sprintf("%s: ", label)
	tp.callback = callback
	tp.origInputMode = termbox.SetInputMode(termbox.InputCurrent)

	termbox.SetInputMode(termbox.InputEsc)

	return tp
}

func (tp *textPrompt) Render(b Box) {
	b.DrawStyledRunes(1, 0, StyledRunesFromString(fmt.Sprintf("%s%s", tp.label, string(tp.text)), gTermStyles.Get(termbox.ColorGreen, termbox.ColorDefault)))
	termbox.SetCursor(b.Left()+1+utf8.RuneCountInString(tp.label)+len(tp.text), b.Top())
}

func (tp *textPrompt) HandleEvent(ks []keyEvent) bool {
	ke := ks[len(ks)-1]

	switch {
	case ke.ctrl && ke.ch == 'c', ke.key == KeyEsc:
		tp.callback("", false)
	case ke.key == KeyEnter:
		tp.callback(string(tp.text), true)
	case ke.key == KeyBackspace:
		if len(tp.text) > 0 {
			tp.text = tp.text[:len(tp.text)-1]
		}
	case ke.key == KeySpace:
		tp.text = append(tp.text, ' ')
	default:
		ch := ke.ch

		if ke.shift {
			ch = unicode.ToUpper(ch)
		}

		tp.text = append(tp.text, ch)
	}

	return true
}

func (tp *textPrompt) Destroy() {
	termbox.SetInputMode(tp.origInputMode)
	termbox.HideCursor()
}
