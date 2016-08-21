package rat

type positionedRune struct {
	ch  rune
	pos BufferPoint
}

func NewPositionedRune(ch rune, pos BufferPoint) PositionedRune {
	return &positionedRune{ch, pos}
}

func (pr *positionedRune) Pos() BufferPoint {
	return pr.pos
}

func (pr *positionedRune) Rune() rune {
	return pr.ch
}
