package rat

import (
	"fmt"
	"unicode/utf8"

	termbox "github.com/nsf/termbox-go"
)

type ConfirmPrompt interface {
	Widget
	Confirm(message string, callback func())
	Clear()
}

type confirmPrompt struct {
	message  string
	callback func()
	box      Box
}

func NewConfirmPrompt() ConfirmPrompt {
	return &confirmPrompt{}
}

func (cp *confirmPrompt) Confirm(message string, callback func()) {
	cp.message = fmt.Sprintf("%s (y/n): ", message)
	cp.callback = callback
	termbox.SetCursor(cp.box.Left()+1+utf8.RuneCountInString(cp.message), cp.box.Top())
}

func (cp *confirmPrompt) Clear() {
	cp.message = ""
	cp.callback = nil
	termbox.HideCursor()
}

func (cp *confirmPrompt) SetBox(b Box) {
	cp.box = b
}

func (cp *confirmPrompt) GetBox() Box {
	return cp.box
}

func (cp *confirmPrompt) Render() {
	cp.box.DrawStyledRunes(1, 0, StyledRunesFromString(cp.message, gTermStyles.Get(termbox.ColorGreen, termbox.ColorDefault)))
}

func (cp *confirmPrompt) HandleEvent(ks []keyEvent) bool {
	if len(cp.message) > 0 && cp.callback != nil {
		switch ks[len(ks)-1] {
		case KeyEventFromString("y"), KeyEventFromString("S-y"):
			cp.callback()
			cp.Clear()
		case KeyEventFromString("n"), KeyEventFromString("S-n"):
			cp.Clear()
		}

		return true
	} else {
		return false
	}
}

func (cp *confirmPrompt) Destroy() {
}
