package escseq

import (
	"io"
	"reflect"
	"strings"
	"testing"
)

func assertEqual(t *testing.T, actual, expected interface{}) {
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected %v, got: %v", expected, actual)
	}
}

type readTokenResult struct {
	token Token
	err   error
}

func testReadToken(t *testing.T, input string, rtt []readTokenResult) {
	l := NewReader(strings.NewReader(input))

	for _, r := range rtt {
		tok, err := l.ReadToken()

		assertEqual(t, tok.Type(), r.token.Type())
		assertEqual(t, tok.Val(), r.token.Val())
		assertEqual(t, err, r.err)
	}
}

func TestReadToken(t *testing.T) {
	// No escape sequences
	testReadToken(t, "abc", []readTokenResult{
		{&token{TokText, "a"}, nil},
		{&token{TokText, "b"}, nil},
		{&token{TokText, "c"}, nil},
		{&token{TokNone, ""}, io.EOF},
	})

	// Simple escape sequence
	testReadToken(t, "\x1bAabc", []readTokenResult{
		{&token{TokEsc, "\x1b"}, nil},
		{&token{TokFe, "A"}, nil},
		{&token{TokText, "a"}, nil},
		{&token{TokText, "b"}, nil},
		{&token{TokText, "c"}, nil},
		{&token{TokNone, ""}, io.EOF},
	})

	// Invalid escape sequence
	testReadToken(t, "\x1baabc", []readTokenResult{
		{&token{TokEsc, "\x1b"}, nil},
		{&token{TokUnknown, "a"}, nil},
		{&token{TokText, "a"}, nil},
		{&token{TokText, "b"}, nil},
		{&token{TokText, "c"}, nil},
		{&token{TokNone, ""}, io.EOF},
	})

	// Empty control sequence
	testReadToken(t, "\x1b[mabc", []readTokenResult{
		{&token{TokEsc, "\x1b"}, nil},
		{&token{TokFe, "["}, nil},
		{&token{TokFinal, "m"}, nil},
		{&token{TokText, "a"}, nil},
		{&token{TokText, "b"}, nil},
		{&token{TokText, "c"}, nil},
		{&token{TokNone, ""}, io.EOF},
	})

	// Control sequence with param substring
	testReadToken(t, "\x1b[31mabc", []readTokenResult{
		{&token{TokEsc, "\x1b"}, nil},
		{&token{TokFe, "["}, nil},
		{&token{TokParamNum, "31"}, nil},
		{&token{TokFinal, "m"}, nil},
		{&token{TokText, "a"}, nil},
		{&token{TokText, "b"}, nil},
		{&token{TokText, "c"}, nil},
		{&token{TokNone, ""}, io.EOF},
	})

	// Control sequence with multiple param substrings
	testReadToken(t, "\x1b[31;1;45mabc", []readTokenResult{
		{&token{TokEsc, "\x1b"}, nil},
		{&token{TokFe, "["}, nil},
		{&token{TokParamNum, "31"}, nil},
		{&token{TokSep, ";"}, nil},
		{&token{TokParamNum, "1"}, nil},
		{&token{TokSep, ";"}, nil},
		{&token{TokParamNum, "45"}, nil},
		{&token{TokFinal, "m"}, nil},
		{&token{TokText, "a"}, nil},
		{&token{TokText, "b"}, nil},
		{&token{TokText, "c"}, nil},
		{&token{TokNone, ""}, io.EOF},
	})

	// Control sequence with separated param substrings
	testReadToken(t, "\x1b[31:125;1:45:11fabc", []readTokenResult{
		{&token{TokEsc, "\x1b"}, nil},
		{&token{TokFe, "["}, nil},
		{&token{TokParamNum, "31"}, nil},
		{&token{TokParamSep, ":"}, nil},
		{&token{TokParamNum, "125"}, nil},
		{&token{TokSep, ";"}, nil},
		{&token{TokParamNum, "1"}, nil},
		{&token{TokParamSep, ":"}, nil},
		{&token{TokParamNum, "45"}, nil},
		{&token{TokParamSep, ":"}, nil},
		{&token{TokParamNum, "11"}, nil},
		{&token{TokFinal, "f"}, nil},
		{&token{TokText, "a"}, nil},
		{&token{TokText, "b"}, nil},
		{&token{TokText, "c"}, nil},
		{&token{TokNone, ""}, io.EOF},
	})

	// Control sequence with reserved param substring bytes
	testReadToken(t, "\x1b[31?;<5mabc", []readTokenResult{
		{&token{TokEsc, "\x1b"}, nil},
		{&token{TokFe, "["}, nil},
		{&token{TokParamNum, "31"}, nil},
		{&token{TokUnknown, "?"}, nil},
		{&token{TokSep, ";"}, nil},
		{&token{TokUnknown, "<"}, nil},
		{&token{TokParamNum, "5"}, nil},
		{&token{TokFinal, "m"}, nil},
		{&token{TokText, "a"}, nil},
		{&token{TokText, "b"}, nil},
		{&token{TokText, "c"}, nil},
		{&token{TokNone, ""}, io.EOF},
	})

	// Control sequence with intermediate bytes
	testReadToken(t, "\x1b[31$.mabc", []readTokenResult{
		{&token{TokEsc, "\x1b"}, nil},
		{&token{TokFe, "["}, nil},
		{&token{TokParamNum, "31"}, nil},
		{&token{TokInter, "$"}, nil},
		{&token{TokInter, "."}, nil},
		{&token{TokFinal, "m"}, nil},
		{&token{TokText, "a"}, nil},
		{&token{TokText, "b"}, nil},
		{&token{TokText, "c"}, nil},
		{&token{TokNone, ""}, io.EOF},
	})

	// Incomplete sequence before EOF
	testReadToken(t, "\x1b[31", []readTokenResult{
		{&token{TokEsc, "\x1b"}, nil},
		{&token{TokFe, "["}, nil},
		{&token{TokParamNum, "31"}, io.EOF},
	})

	// Control sequence with private param string
	testReadToken(t, "\x1b[?31>41F", []readTokenResult{
		{&token{TokEsc, "\x1b"}, nil},
		{&token{TokFe, "["}, nil},
		{&token{TokPrivParam, "?31>41"}, nil},
		{&token{TokFinal, "F"}, nil},
		{&token{TokNone, ""}, io.EOF},
	})
}
