package rat

import termbox "github.com/nsf/termbox-go"

type Box interface {
	Left() int
	Top() int
	Width() int
	Height() int
	DrawStyledRune(x, y int, sr StyledRune)
	DrawStyledRunes(x, y int, runes []StyledRune)
	Fill(sr StyledRune)
}

type box struct {
	left   int
	top    int
	width  int
	height int
}

func NewBox(left, top, width, height int) Box {
	return &box{left, top, width, height}
}

func (b *box) Left() int {
	return b.left
}

func (b *box) Top() int {
	return b.top
}

func (b *box) Width() int {
	return b.width
}

func (b *box) Height() int {
	return b.height
}

func (b *box) DrawStyledRune(x, y int, sr StyledRune) {
	if x >= 0 && x < b.width && y >= 0 && y < b.height {
		termbox.SetCell(b.left+x, b.top+y, sr.Rune(), sr.Fg(), sr.Bg())
	}
}

func (b *box) DrawStyledRunes(x, y int, runes []StyledRune) {
	offset := 0
	var r rune
	var adv int

	for _, sr := range runes {
		r = sr.Rune()

		if r != '\t' && r != '\n' {
			b.DrawStyledRune(x+offset, y, sr)
		}

		if r == '\t' {
			adv = 8 - offset%8
		} else {
			adv = 1
		}

		offset = offset + adv
	}
}

func (b *box) Fill(sr StyledRune) {
	for y := 0; y < b.height; y++ {
		for x := 0; x < b.width; x++ {
			b.DrawStyledRune(x, y, sr)
		}
	}
}
