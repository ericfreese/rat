package rat

type bufferPoint struct {
	line int
	col  int
}

func NewBufferPoint(line, col int) BufferPoint {
	return &bufferPoint{line, col}
}

func (bp *bufferPoint) Line() int {
	return bp.line
}

func (bp *bufferPoint) Col() int {
	return bp.col
}
