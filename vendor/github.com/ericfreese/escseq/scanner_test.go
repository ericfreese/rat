package escseq

import (
	"strings"
	"testing"
)

func TestUnreadToken(t *testing.T) {
	s := NewScanner(strings.NewReader("abc"))

	s.ReadToken()               // Read "a"
	tok1, err1 := s.ReadToken() // Read "b"
	err := s.UnreadToken()      // Unread "b"
	tok2, err2 := s.ReadToken() // Read "b"

	assertEqual(t, tok1.Type(), tok2.Type())
	assertEqual(t, tok1.Val(), tok2.Val())
	assertEqual(t, err1, err2)
	assertEqual(t, err, nil)

	// Error from two sequential unreads
	err = s.UnreadToken() // Unread "b"
	assertEqual(t, err, nil)
	err = s.UnreadToken() // Unread "b" again (causes error)
	assertEqual(t, err, ErrInvalidUnreadToken)
}
