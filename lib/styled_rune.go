package rat

import "unicode/utf8"

type StyledRune interface {
	Rune() rune
	TermStyle
}

type styledRune struct {
	ch rune
	TermStyle
}

func NewStyledRune(ch rune, ts TermStyle) StyledRune {
	return &styledRune{ch, ts}
}

func StyledRunesFromString(str string, ts TermStyle) []StyledRune {
	runes := make([]StyledRune, 0, utf8.RuneCountInString(str))

	for _, r := range str {
		runes = append(runes, NewStyledRune(r, ts))
	}

	return runes
}

func (sr *styledRune) Rune() rune {
	return sr.ch
}
