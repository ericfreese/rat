package escseq

import (
	"errors"
	"io"
)

var ErrInvalidUnreadToken = errors.New("escseq: invalid use of UnreadToken")

// The Scanner interface adds the UnreadToken method to the basic
// ReadToken method.
//
// UnreadToken causes the next call to ReadToken to return the same
// Token as the previous call to ReadToken. It is an error to call
// UnreadToken twice without calling ReadToken
type Scanner interface {
	Reader
	UnreadToken() error
}

type scanner struct {
	rd        Reader
	lastToken Token
	lastErr   error
	readLast  bool
}

// Returns a Scanner that reads tokens from the provided io.Reader.
func NewScanner(rd io.Reader) Scanner {
	s := &scanner{}
	s.rd = NewReader(rd)
	return s
}

func (s *scanner) ReadToken() (Token, error) {
	if s.readLast {
		s.readLast = false
		return s.lastToken, s.lastErr
	}

	s.lastToken, s.lastErr = s.rd.ReadToken()

	return s.lastToken, s.lastErr
}

func (s *scanner) UnreadToken() error {
	if s.readLast {
		return ErrInvalidUnreadToken
	} else {
		s.readLast = true
		return nil
	}
}
