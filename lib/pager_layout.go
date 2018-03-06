package rat

import (
	"fmt"

	termbox "github.com/nsf/termbox-go"
)

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
