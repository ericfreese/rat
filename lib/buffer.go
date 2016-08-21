package rat

import (
	"io"
	"sync"
	"unicode/utf8"
)

type buffer struct {
	lines           [][]StyledRune
	annotations     []Annotation
	appendListeners []chan PositionedRune
	appendLock      sync.Mutex
	annotationLock  sync.Mutex
	stopped         chan bool
	finished        bool
}

func NewBuffer(srr StyledRuneReader, initParsers func() []Annotator) Buffer {
	b := &buffer{}

	b.lines = append(make([][]StyledRune, 0, 256), make([]StyledRune, 0, 128))
	b.appendListeners = make([]chan PositionedRune, 0, 16)
	b.stopped = make(chan bool)

	go b.appendFrom(srr)

	for _, ap := range initParsers() {
		go b.annotateWith(ap)
	}

	return b
}

func (b *buffer) LineRange(start int, numLines int) [][]StyledRune {
	if start > len(b.lines)-1 {
		return nil
	} else if start+numLines < len(b.lines) {
		return b.lines[start : start+numLines]
	} else {
		return b.lines[start:]
	}
}

func (b *buffer) NumLines() int {
	return len(b.lines)
}

func (b *buffer) NumAnnotations() int {
	return len(b.annotations)
}

func (b *buffer) AnnotationsForLine(line int) []Annotation {
	annotations := make([]Annotation, 0, 8)

	for _, a := range b.annotations {
		if a.Start().Line() <= line && a.End().Line() >= line {
			annotations = append(annotations, a)
		}
	}

	return annotations
}

func (b *buffer) stop() {
	select {
	case <-b.stopped:
	default:
		close(b.stopped)
	}
}

func (b *buffer) Destroy() {
	b.stop()
}

func (b *buffer) Lock() {
	b.appendLock.Lock()
	b.annotationLock.Lock()
}

func (b *buffer) Unlock() {
	b.appendLock.Unlock()
	b.annotationLock.Unlock()
}

func (b *buffer) NextPositionedRune(bp BufferPoint) (PositionedRune, error) {
	var line, col int
	var next BufferPoint

	if bp == nil {
		line = 0
		col = -1
	} else {
		line = bp.Line()
		col = bp.Col()
	}

	b.appendLock.Lock()

	if col+1 < len(b.lines[line]) {
		next = NewBufferPoint(line, col+1)
	} else if line+1 < len(b.lines) && len(b.lines[line+1]) > 0 {
		next = NewBufferPoint(line+1, 0)
	}

	if next != nil {
		defer b.appendLock.Unlock()
		return NewPositionedRune(b.lines[next.Line()][next.Col()].Rune(), next), nil
	} else {
		select {
		case <-b.stopped:
			b.appendLock.Unlock()
			return nil, io.EOF
		default:
			nextAppend := make(chan PositionedRune, 1)
			b.appendListeners = append(b.appendListeners, nextAppend)
			b.appendLock.Unlock()

			if pr, ok := <-nextAppend; ok {
				close(nextAppend)
				return pr, nil
			} else {
				return nil, io.EOF
			}
		}
	}
}

func (b *buffer) annotateWith(ap Annotator) {
	for a := range ap.Annotate(NewBufferReader(b)) {
		b.annotationLock.Lock()
		b.annotations = append(b.annotations, a)
		b.annotationLock.Unlock()
	}
}

func (b *buffer) appendFrom(srr StyledRuneReader) {
	for {
		select {
		case <-b.stopped:
			return
		default:
			sr, err := srr.ReadStyledRune()

			if sr.Rune() != utf8.RuneError {
				b.append(sr)
			}

			if err != nil {
				b.appendLock.Lock()
				defer b.appendLock.Unlock()

				for _, l := range b.appendListeners {
					close(l)
				}

				b.stop()

				return
			}
		}

	}
}

func (b *buffer) append(sr StyledRune) {
	b.appendLock.Lock()
	defer b.appendLock.Unlock()

	if len(b.appendListeners) > 0 {
		curLine := len(b.lines) - 1
		curCol := len(b.lines[curLine])
		pr := NewPositionedRune(sr.Rune(), NewBufferPoint(curLine, curCol))

		for _, l := range b.appendListeners {
			l <- pr
		}

		b.appendListeners = b.appendListeners[0:0]
	}

	b.lines[len(b.lines)-1] = append(b.lines[len(b.lines)-1], sr)

	if sr.Rune() == '\n' {
		b.lines = append(b.lines, make([]StyledRune, 0, 128))
	}
}
