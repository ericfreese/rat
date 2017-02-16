package rat

import (
	"sync"

	termbox "github.com/nsf/termbox-go"
)

var gTermStyles = NewTermStyles()

type TermStyle interface {
	Fg() termbox.Attribute
	Bg() termbox.Attribute
}

type TermStyles interface {
	Get(termbox.Attribute, termbox.Attribute) TermStyle
	Default() TermStyle
}

type termStyle struct {
	fg termbox.Attribute
	bg termbox.Attribute
}

func (sp *termStyle) Fg() termbox.Attribute {
	return sp.fg
}

func (sp *termStyle) Bg() termbox.Attribute {
	return sp.bg
}

type termStyles struct {
	cache map[tsCacheKey]TermStyle
	sync.Mutex
}

type tsCacheKey struct {
	fg termbox.Attribute
	bg termbox.Attribute
}

func NewTermStyles() TermStyles {
	ts := &termStyles{}
	ts.cache = make(map[tsCacheKey]TermStyle)
	return ts
}

func (ts *termStyles) Get(fg, bg termbox.Attribute) TermStyle {
	ts.Lock()
	defer ts.Unlock()

	k := tsCacheKey{fg, bg}

	if s, ok := ts.cache[k]; ok {
		return s
	} else {
		s := &termStyle{fg, bg}
		ts.cache[k] = s
		return s
	}
}

func (ts *termStyles) Default() TermStyle {
	return ts.Get(termbox.ColorDefault, termbox.ColorDefault)
}
