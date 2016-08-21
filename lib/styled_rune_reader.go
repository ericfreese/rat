package rat

import (
	"io"
	"strconv"
	"unicode/utf8"

	"github.com/ericfreese/escseq"
	termbox "github.com/nsf/termbox-go"
)

type styledRuneReader struct {
	reader  io.Reader
	scanner escseq.Scanner
	out     chan StyledRune
	fg      termbox.Attribute
	bg      termbox.Attribute
}

func NewStyledRuneReader(reader io.Reader) StyledRuneReader {
	srr := &styledRuneReader{}

	srr.scanner = escseq.NewScanner(reader)

	return srr
}

func (srr *styledRuneReader) ReadStyledRune() (StyledRune, error) {
	t, err := srr.scanner.ReadToken()

	switch t.Type() {
	case escseq.TokText:
		r, _ := utf8.DecodeRuneInString(t.Val())
		return srr.styleRune(r), err
	case escseq.TokEsc:
		srr.scanner.UnreadToken()
		srr.processEscSeq()
		return srr.ReadStyledRune()
	case escseq.TokNone:
		return srr.styleRune(utf8.RuneError), err
	default:
		panic("unexpected token type")
	}
}

func (srr *styledRuneReader) processEscSeq() {
	toks := make([]escseq.Token, 0, 16)
	numSgr := 0

	for {
		t, err := srr.scanner.ReadToken()

		if t.Type() != escseq.TokNone {
			toks = append(toks, t)
		}

		if t.Type() == escseq.TokParamNum {
			numSgr++
		}

		if (t.Type() == escseq.TokFe && t.Val() != "[") || t.Type() == escseq.TokFinal {
			break
		}

		if err != nil {
			break
		}
	}

	sgr := make([]int, 0, numSgr)

	for _, t := range toks {
		if t.Type() == escseq.TokFe && t.Val() != "[" ||
			t.Type() == escseq.TokInter ||
			t.Type() == escseq.TokFinal && t.Val() != "m" {
			return
		}

		if t.Type() == escseq.TokParamNum {
			i, _ := strconv.Atoi(t.Val())
			sgr = append(sgr, i)
		}
	}

	srr.processSGRParams(sgr)
}

func (srr *styledRuneReader) processSGRParams(sgr []int) {
	n := len(sgr)

	if n == 0 {
		srr.fg = termbox.ColorDefault
		srr.bg = termbox.ColorDefault
		return
	}

	for i := 0; i < n; i++ {
		p := sgr[i]

		switch {
		case p == 0:
			srr.fg = termbox.ColorDefault
			srr.bg = termbox.ColorDefault
		case p == 1:
			srr.fg = srr.fg | termbox.AttrBold
		case p == 4:
			srr.fg = srr.fg | termbox.AttrUnderline
		case p == 7:
			srr.fg = srr.fg | termbox.AttrReverse
		case p == 27:
			srr.fg = srr.fg &^ termbox.AttrReverse
		case p >= 30 && p <= 37:
			srr.fg = srr.fg | termbox.Attribute(p-29)
		case p == 38:
			if n-i >= 2 && sgr[i+1] == 5 {
				srr.fg = srr.fg | termbox.Attribute(sgr[i+2]+1)
				i = i + 2
			}
		case p == 39:
			srr.fg = termbox.ColorDefault
		case p >= 40 && p <= 47:
			srr.bg = srr.bg | termbox.Attribute(p-39)
		case p == 48:
			if n-i >= 2 && sgr[i+1] == 5 {
				srr.bg = srr.bg | termbox.Attribute(sgr[i+2]+1)
				i = i + 2
			}
		case p == 49:
			srr.bg = termbox.ColorDefault
		}
	}
}

func (srr *styledRuneReader) styleRune(r rune) StyledRune {
	return NewStyledRune(r, srr.fg, srr.bg)
}
