package rat

import (
	"bufio"
	"io"
	"sync"

	termbox "github.com/nsf/termbox-go"
)

type Buffer interface {
	sync.Locker
	AnnotateWith(Annotator)
	AnnotationsForLine(line int) []Annotation
	NumAnnotations() int
	NumLines() int
	StyledLines(start int, numLines int) [][]StyledRune
	io.Closer
}

type buffer struct {
	stream      Stream
	highlights  Highlights
	annotations Annotations
	lines       []Line

	sync.Mutex
}

func NewBuffer(rd io.Reader) Buffer {
	b := &buffer{}

	b.stream = NewStream()
	b.highlights = NewHighlights()
	b.annotations = NewAnnotations()
	b.lines = make([]Line, 1, 64)
	b.lines[0] = NewLine(0, 0)

	go b.processTokens(NewScanner(rd))

	return b
}

func (b *buffer) AnnotateWith(annotator Annotator) {
	for a := range annotator.Annotate(b.stream.NewReader()) {
		b.Lock()
		b.annotations.Add(a)
		b.Unlock()
	}
}

func (b *buffer) AnnotationsForLine(l int) []Annotation {
	if l < 0 {
		panic("invalid line index")
	}

	if l >= len(b.lines) {
		return []Annotation{}
	}

	return b.annotations.Intersecting(b.lines[l])
}

func (b *buffer) NumAnnotations() int {
	return b.annotations.Len()
}

func (b *buffer) NumLines() int {
	return len(b.lines)
}

func (b *buffer) StyledLines(start, numLines int) [][]StyledRune {
	if start < 0 {
		panic("invalid line index")
	}

	if start >= len(b.lines) {
		numLines = 0
	} else if start+numLines > len(b.lines) {
		numLines = len(b.lines) - start
	}

	styledLines := make([][]StyledRune, 0, numLines)

	var lines []Line
	if numLines > 0 {
		lines = b.lines[start : start+numLines]
	} else {
		lines = make([]Line, 0, 0)
	}

	for i, line := range lines {
		lineString := string(b.stream.Bytes()[line.Start():line.End()])

		styledLines = append(styledLines, make([]StyledRune, 0, len(lineString)))

		offset := line.Start()
		var sr StyledRune
		for _, r := range lineString {
			if h := b.highlights.AtPoint(offset); h != nil {
				sr = NewStyledRune(r, h)
			} else {
				sr = NewStyledRune(r, gTermStyles.Default())
			}

			styledLines[i] = append(styledLines[i], sr)

			offset = offset + len(string(r))
		}
	}

	return styledLines
}

func (b *buffer) Close() error {
	return b.stream.Close()
}

func (b *buffer) processTokens(tr TokenReader) {
	w := bufio.NewWriter(b.stream)

	var (
		offset int
		t      Token
		err    error
		n      int
	)

	for {
		t, err = tr.ReadToken()
		if err != nil {
			break
		}

		b.Lock()

		if len(t.Val()) > 0 {
			n, err = w.WriteString(string(t.Val()))
			w.Flush()

			offset = offset + n

			b.lines[len(b.lines)-1].SetEnd(offset)
		}

		if t.Type() == TokTermStyle {
			sp := t.TermStyle()

			b.highlights.End(offset)

			if sp.Fg() != termbox.ColorDefault || sp.Bg() != termbox.ColorDefault {
				b.highlights.Start(offset, t.TermStyle())
			}
		}

		if t.Type() == TokNewLine {
			b.lines = append(b.lines, NewLine(b.lines[len(b.lines)-1].End(), offset))
		}

		b.Unlock()
	}

	b.Close()
}
