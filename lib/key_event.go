package rat

import (
	"strings"
	"unicode"
	"unicode/utf8"

	termbox "github.com/nsf/termbox-go"
)

type KeyConst termbox.Key

const (
	KeyF1         KeyConst = KeyConst(termbox.KeyF1)
	KeyF2         KeyConst = KeyConst(termbox.KeyF2)
	KeyF3         KeyConst = KeyConst(termbox.KeyF3)
	KeyF4         KeyConst = KeyConst(termbox.KeyF4)
	KeyF5         KeyConst = KeyConst(termbox.KeyF5)
	KeyF6         KeyConst = KeyConst(termbox.KeyF6)
	KeyF7         KeyConst = KeyConst(termbox.KeyF7)
	KeyF8         KeyConst = KeyConst(termbox.KeyF8)
	KeyF9         KeyConst = KeyConst(termbox.KeyF9)
	KeyF10        KeyConst = KeyConst(termbox.KeyF10)
	KeyF11        KeyConst = KeyConst(termbox.KeyF11)
	KeyF12        KeyConst = KeyConst(termbox.KeyF12)
	KeyInsert     KeyConst = KeyConst(termbox.KeyInsert)
	KeyDelete     KeyConst = KeyConst(termbox.KeyDelete)
	KeyHome       KeyConst = KeyConst(termbox.KeyHome)
	KeyEnd        KeyConst = KeyConst(termbox.KeyEnd)
	KeyPgup       KeyConst = KeyConst(termbox.KeyPgup)
	KeyPgdn       KeyConst = KeyConst(termbox.KeyPgdn)
	KeyArrowUp    KeyConst = KeyConst(termbox.KeyArrowUp)
	KeyArrowDown  KeyConst = KeyConst(termbox.KeyArrowDown)
	KeyArrowLeft  KeyConst = KeyConst(termbox.KeyArrowLeft)
	KeyArrowRight KeyConst = KeyConst(termbox.KeyArrowRight)
	KeyTab        KeyConst = KeyConst(termbox.KeyTab)
	KeyBackspace  KeyConst = KeyConst(termbox.KeyBackspace)
	KeyEnter      KeyConst = KeyConst(termbox.KeyEnter)
	KeyEsc        KeyConst = KeyConst(termbox.KeyEsc)
	KeySpace      KeyConst = KeyConst(termbox.KeySpace)
)

var keyStrings = map[string]KeyConst{
	"f1":        KeyF1,
	"f2":        KeyF2,
	"f3":        KeyF3,
	"f4":        KeyF4,
	"f5":        KeyF5,
	"f6":        KeyF6,
	"f7":        KeyF7,
	"f8":        KeyF8,
	"f9":        KeyF9,
	"f10":       KeyF10,
	"f11":       KeyF11,
	"f12":       KeyF12,
	"insert":    KeyInsert,
	"delete":    KeyDelete,
	"home":      KeyHome,
	"end":       KeyEnd,
	"pgup":      KeyPgup,
	"pgdn":      KeyPgdn,
	"up":        KeyArrowUp,
	"down":      KeyArrowDown,
	"left":      KeyArrowLeft,
	"right":     KeyArrowRight,
	"tab":       KeyTab,
	"backspace": KeyBackspace,
	"enter":     KeyEnter,
	"esc":       KeyEsc,
	"space":     KeySpace,
}

var tbKeys = map[termbox.Key]KeyConst{
	termbox.KeyF1:         KeyF1,
	termbox.KeyF2:         KeyF2,
	termbox.KeyF3:         KeyF3,
	termbox.KeyF4:         KeyF4,
	termbox.KeyF5:         KeyF5,
	termbox.KeyF6:         KeyF6,
	termbox.KeyF7:         KeyF7,
	termbox.KeyF8:         KeyF8,
	termbox.KeyF9:         KeyF9,
	termbox.KeyF10:        KeyF10,
	termbox.KeyF11:        KeyF11,
	termbox.KeyF12:        KeyF12,
	termbox.KeyInsert:     KeyInsert,
	termbox.KeyDelete:     KeyDelete,
	termbox.KeyHome:       KeyHome,
	termbox.KeyEnd:        KeyEnd,
	termbox.KeyPgup:       KeyPgup,
	termbox.KeyPgdn:       KeyPgdn,
	termbox.KeyArrowUp:    KeyArrowUp,
	termbox.KeyArrowDown:  KeyArrowDown,
	termbox.KeyArrowLeft:  KeyArrowLeft,
	termbox.KeyArrowRight: KeyArrowRight,
	termbox.KeyTab:        KeyTab,
	termbox.KeyBackspace:  KeyBackspace,
	termbox.KeyBackspace2: KeyBackspace,
	termbox.KeyEnter:      KeyEnter,
	termbox.KeyEsc:        KeyEsc,
	termbox.KeySpace:      KeySpace,
}

var tbCtrlRunes = map[termbox.Key]rune{
	termbox.KeyCtrlA: 'a',
	termbox.KeyCtrlB: 'b',
	termbox.KeyCtrlC: 'c',
	termbox.KeyCtrlD: 'd',
	termbox.KeyCtrlE: 'e',
	termbox.KeyCtrlF: 'f',
	termbox.KeyCtrlG: 'g',
	termbox.KeyCtrlH: 'h',
	termbox.KeyCtrlI: 'i',
	termbox.KeyCtrlJ: 'j',
	termbox.KeyCtrlK: 'k',
	termbox.KeyCtrlL: 'l',
	termbox.KeyCtrlM: 'm',
	termbox.KeyCtrlN: 'n',
	termbox.KeyCtrlO: 'o',
	termbox.KeyCtrlP: 'p',
	termbox.KeyCtrlQ: 'q',
	termbox.KeyCtrlR: 'r',
	termbox.KeyCtrlS: 's',
	termbox.KeyCtrlT: 't',
	termbox.KeyCtrlU: 'u',
	termbox.KeyCtrlV: 'v',
	termbox.KeyCtrlW: 'w',
	termbox.KeyCtrlX: 'x',
	termbox.KeyCtrlY: 'y',
	termbox.KeyCtrlZ: 'z',
	termbox.KeyCtrl3: '3',
	termbox.KeyCtrl4: '4',
	termbox.KeyCtrl5: '5',
	termbox.KeyCtrl6: '6',
	termbox.KeyCtrl7: '7',
	termbox.KeyCtrl8: '8',
}

type keyEvent struct {
	ctrl  bool
	meta  bool
	shift bool
	key   KeyConst
	ch    rune
}

func KeyEventFromTBEvent(tbev *termbox.Event) keyEvent {
	ke := keyEvent{}

	ke.meta = tbev.Mod == termbox.ModAlt

	if k, ok := tbKeys[tbev.Key]; ok {
		ke.key = k
	} else if ch, ok := tbCtrlRunes[tbev.Key]; ok {
		ke.ctrl = true
		ke.ch = ch
	} else if tbev.Ch != 0 {
		ke.shift = unicode.IsUpper(tbev.Ch)
		ke.ch = unicode.ToLower(tbev.Ch)
	}

	return ke
}

func KeyEventFromString(str string) keyEvent {
	ke := keyEvent{}

	modsParsed := false
	for !modsParsed {
		switch {
		case strings.HasPrefix(str, "C-"):
			ke.ctrl = true
			str = str[2:]
		case strings.HasPrefix(str, "M-"):
			ke.meta = true
			str = str[2:]
		case strings.HasPrefix(str, "S-"):
			ke.shift = true
			str = str[2:]
		default:
			modsParsed = true
		}
	}

	if k, ok := keyStrings[str]; ok {
		ke.key = k
	} else if utf8.RuneCountInString(str) == 1 {
		r, _ := utf8.DecodeRuneInString(strings.ToLower(str))
		ke.ch = r
	}

	return ke
}

// TODO: Rewrite to allow commas in here
func KeySequenceFromString(str string) []keyEvent {
	keyStrings := strings.Split(str, ",")
	ks := make([]keyEvent, 0, len(keyStrings))

	for _, keyStr := range keyStrings {
		ks = append(ks, KeyEventFromString(keyStr))
	}

	return ks
}
