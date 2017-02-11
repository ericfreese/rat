package escseq

import (
	"bufio"
	"bytes"
	"io"
)

type readerState uint8

const (
	rsDefault readerState = iota
	rdEsc
	rdCSParam
	rdCSInter
)

// The Reader interface wraps the ReadToken method.
//
// ReadToken reads a single Token from the input stream. It will
// return the Token and any errors it encountered while reading the
// Token. If no Token can be read, it will return a Token with type
// TokNone. If an unexpected input is found, it will return a Token
// with type TokUnknown.
type Reader interface {
	ReadToken() (Token, error)
}

type reader struct {
	lastState readerState
	state     readerState
	rs        io.RuneScanner
}

// Returns a Reader that reads tokens from the provided io.Reader.
func NewReader(rd io.Reader) Reader {
	return &reader{rsDefault, rsDefault, bufio.NewReader(rd)}
}

func (rd *reader) ReadToken() (Token, error) {
	r, _, err := rd.rs.ReadRune()

	if err != nil {
		return &token{TokNone, ""}, err
	}

	switch rd.state {
	case rsDefault:
		if r == 0x1b {
			rd.setState(rdEsc)
			return &token{TokEsc, string(r)}, nil
		} else {
			return &token{TokText, string(r)}, nil
		}
	case rdEsc:
		if r >= 0x40 && r <= 0x5f {
			if r == '[' {
				rd.setState(rdCSParam)
			} else {
				rd.setState(rsDefault)
			}

			return &token{TokFe, string(r)}, nil
		}
	case rdCSParam:
		switch {
		case r >= 0x40 && r <= 0x7e:
			rd.setState(rsDefault)
			return &token{TokFinal, string(r)}, nil
		case rd.lastState == rdEsc && r >= 0x3c && r <= 0x3f:
			rd.rs.UnreadRune()
			return rd.readCSPrivParam()
		case r >= 0x30 && r <= 0x3f:
			rd.setState(rdCSParam)

			switch {
			case r >= 0x3c && r <= 0x3f:
				return &token{TokUnknown, string(r)}, nil
			case r == ':':
				return &token{TokParamSep, string(r)}, nil
			case r == ';':
				return &token{TokSep, string(r)}, nil
			default:
				rd.rs.UnreadRune()
				return rd.readCSParamNum()
			}
		case r >= 0x20 && r <= 0x2f:
			rd.setState(rdCSInter)
			return &token{TokInter, string(r)}, nil
		}
	case rdCSInter:
		switch {
		case r >= 0x40 && r <= 0x7e:
			rd.setState(rsDefault)
			return &token{TokFinal, string(r)}, nil
		case r >= 0x20 && r <= 0x2f:
			rd.setState(rdCSInter)
			return &token{TokInter, string(r)}, nil
		}
	}

	rd.setState(rsDefault)
	return &token{TokUnknown, string(r)}, nil
}

func (rd *reader) setState(s readerState) {
	rd.lastState = rd.state
	rd.state = s
}

func (rd *reader) readCSParamNum() (Token, error) {
	var (
		buf bytes.Buffer
		r   rune
		err error
	)

	for {
		r, _, err = rd.rs.ReadRune()

		if err != nil {
			break
		}

		if r >= 0x30 && r <= 0x39 {
			buf.WriteRune(r)
		} else {
			rd.rs.UnreadRune()
			break
		}
	}

	return &token{TokParamNum, buf.String()}, err
}

func (rd *reader) readCSPrivParam() (Token, error) {
	var (
		buf bytes.Buffer
		r   rune
		err error
	)

	for {
		r, _, err = rd.rs.ReadRune()

		if err != nil {
			break
		}

		if r >= 0x30 && r <= 0x3f {
			buf.WriteRune(r)
		} else {
			rd.rs.UnreadRune()
			break
		}
	}

	return &token{TokPrivParam, buf.String()}, err
}
