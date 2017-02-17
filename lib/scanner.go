package rat

import (
	"io"
	"strconv"

	"github.com/ericfreese/escseq"
	termbox "github.com/nsf/termbox-go"
)

type TokenType uint8

const (
	TokNone TokenType = iota
	TokText
	TokTermStyle
	TokNewLine
)

type Token interface {
	Type() TokenType
	Val() []byte
	TermStyle() TermStyle
}

type TokenReader interface {
	ReadToken() (Token, error)
}

type token struct {
	t   TokenType
	val []byte
	sp  TermStyle
}

func (t *token) Type() TokenType {
	return t.t
}

func (t *token) Val() []byte {
	return t.val
}

func (t *token) TermStyle() TermStyle {
	return t.sp
}

type scanner struct {
	es escseq.Scanner
	fg termbox.Attribute
	bg termbox.Attribute
}

func NewScanner(rd io.Reader) TokenReader {
	s := &scanner{}
	s.es = escseq.NewScanner(rd)
	return s
}

func (s *scanner) ReadToken() (Token, error) {
	t, err := s.es.ReadToken()

	switch t.Type() {
	case escseq.TokText:
		if t.Val() == "\n" {
			return &token{TokNewLine, []byte(t.Val()), nil}, err
		} else {
			return &token{TokText, []byte(t.Val()), nil}, err
		}
	case escseq.TokEsc:
		s.es.UnreadToken()
		if termStyle := s.scanTermStyle(); termStyle != nil {
			return &token{TokTermStyle, []byte{}, termStyle}, err
		} else {
			return &token{TokNone, []byte{}, nil}, err
		}
	default:
		return &token{TokNone, []byte{}, nil}, err
	}
}

func (s *scanner) scanTermStyle() TermStyle {
	toks := make([]escseq.Token, 0, 16)
	numSgr := 0

	for {
		t, err := s.es.ReadToken()

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
			return nil
		}

		if t.Type() == escseq.TokParamNum {
			i, _ := strconv.Atoi(t.Val())
			sgr = append(sgr, i)
		}
	}

	return s.buildTermStyle(sgr)
}

func (s *scanner) buildTermStyle(sgr []int) TermStyle {
	n := len(sgr)

	if n == 0 {
		s.fg = termbox.ColorDefault
		s.bg = termbox.ColorDefault
		return gTermStyles.Get(s.fg, s.bg)
	}

	for i := 0; i < n; i++ {
		p := sgr[i]

		switch {
		case p == 0:
			s.fg = termbox.ColorDefault
			s.bg = termbox.ColorDefault
		case p == 1:
			s.fg = s.fg | termbox.AttrBold
		case p == 4:
			s.fg = s.fg | termbox.AttrUnderline
		case p == 7:
			s.fg = s.fg | termbox.AttrReverse
		case p == 27:
			s.fg = s.fg &^ termbox.AttrReverse
		case p >= 30 && p <= 37:
			s.fg = s.fg | termbox.Attribute(p-29)
		case p == 38:
			if n-i > 2 && sgr[i+1] == 5 {
				s.fg = s.fg | termbox.Attribute(sgr[i+2]+1)
				i = i + 2
			}
		case p == 39:
			s.fg = termbox.ColorDefault
		case p >= 40 && p <= 47:
			s.bg = s.bg | termbox.Attribute(p-39)
		case p == 48:
			if n-i > 2 && sgr[i+1] == 5 {
				s.bg = s.bg | termbox.Attribute(sgr[i+2]+1)
				i = i + 2
			}
		case p == 49:
			s.bg = termbox.ColorDefault
		}
	}

	return gTermStyles.Get(s.fg, s.bg)
}
