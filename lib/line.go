package rat

type Line interface {
	Start() int
	End() int
	SetEnd(int)
}

type line struct {
	start int
	end   int
}

func NewLine(start, end int) Line {
	return &line{start, end}
}

func (l *line) Start() int {
	return l.start
}

func (l *line) End() int {
	return l.end
}

func (l *line) SetEnd(p int) {
	l.end = p
}
