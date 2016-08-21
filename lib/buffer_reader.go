package rat

import "io"

type bufferReader struct {
	b   Buffer
	pos BufferPoint
}

func NewBufferReader(b Buffer) BufferReader {
	return &bufferReader{b, nil}
}

func (br *bufferReader) ReadPositionedRune() (PositionedRune, error) {
	next, err := br.b.NextPositionedRune(br.pos)

	if err == io.EOF {
		return nil, err
	} else {
		br.pos = next.Pos()
		return next, nil
	}
}
