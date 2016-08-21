package rat

// TODO: Decouple from termbox
import (
	"unicode/utf8"

	"github.com/nsf/termbox-go"
)

type styledRune struct {
	ch rune
	fg termbox.Attribute
	bg termbox.Attribute
}

func NewStyledRune(ch rune, fg termbox.Attribute, bg termbox.Attribute) StyledRune {
	return &styledRune{ch, fg, bg}
}

func StyledRunesFromString(str string, fg termbox.Attribute, bg termbox.Attribute) []StyledRune {
	runes := make([]StyledRune, 0, utf8.RuneCountInString(str))

	for _, r := range str {
		runes = append(runes, NewStyledRune(r, fg, bg))
	}

	return runes
}

func (sr *styledRune) Fg() termbox.Attribute {
	return sr.fg
}

func (sr *styledRune) Bg() termbox.Attribute {
	return sr.bg
}

func (sr *styledRune) Rune() rune {
	return sr.ch
}
