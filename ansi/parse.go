package ansi

import (
	"github.com/muesli/termenv"
)

type SequenceType int
type SGRType int

const (
	SequenceTypeNone SequenceType = iota
	SequenceTypeCursorUp
	SequenceTypeCursorDown
	SequenceTypeCursorForward
	SequenceTypeCursorBack
	SequenceTypeCursorNextLine
	SequenceTypeCursorPreviousLine
	SequenceTypeCursorHorizontal
	SequenceTypeCursorPosition
	SequenceTypeEraseDisplay
	SequenceTypeEraseLine
	SequenceTypeScrollUp
	SequenceTypeScrollDown
	SequenceTypeSaveCursorPosition
	SequenceTypeSGR
	SequenceTypeShowCursor
	SequenceTypeHideCursor
	SequenceTypeUnimplemented
)

const (
	SGRTypeReset          SGRType = 0
	SGRTypeBold           SGRType = 1
	SGRTypeUnsetBold      SGRType = 22
	SGRTypeItalic         SGRType = 3
	SGRTypeUnsetItalic    SGRType = 23
	SGRTypeUnderline      SGRType = 4
	SGRTypeUnsetUnderline SGRType = 24
	SGRTypeFgColor        SGRType = 38
	SGRTypeBgColor        SGRType = 48
	SGRTypeFgDefaultColor SGRType = 39
	SGRTypeBgDefaultColor SGRType = 49
)

const (
	// States for the internal state machine
	stateStart = 0
	stateParam = 1
	stateInter = 2
)

var (
	csi = []rune(termenv.CSI)
)

type Cursor struct {
	SeqType SequenceType

	runes      *[]rune
	start, end int
}

func (c *Cursor) Len() int {
	diff := c.end - c.start
	if diff < 0 {
		return 0
	}
	return diff
}

func (c *Cursor) View() []rune {
	if c.Len() == 0 {
		return nil
	}

	return (*c.runes)[c.start:c.end]
}

func (c *Cursor) Arg1() byte {
	if c.Len() == 0 {
		return 0
	}
	id, _ := parseNum(c.View())
	return byte(id)
}

func (c *Cursor) Arg2() (byte, byte) {
	if c.Len() == 0 {
		return 0, 0
	}

	a1, off := parseNum(c.View())
	a2, _ := parseNum(c.View()[off+1:])

	return byte(a1), byte(a2)
}

func (c *Cursor) VisitSGR(fn func(t SGRType, params []byte)) {
	if c.SeqType != SequenceTypeSGR {
		return
	}

	// No parameters at all in ESC[m acts like a 0 reset code
	if c.Len() == 0 {
		fn(SGRTypeReset, []byte{})
		return
	}

	params := make([]byte, 0, 3)

	// We have parameters, so we parse them
	view := c.View()
	for i := 0; i < len(view); i++ {
		if id, off := parseNum(view[i:]); off > 0 {
			i += off - 1
			t := SGRType(id)

			switch t {
			case SGRTypeFgColor, SGRTypeBgColor:
				if ok, trueColor, r, g, b, off := parseColor(view[i+2:]); ok {
					i += off + 2

					if trueColor {
						params = append(params[:0], byte(r), byte(g), byte(b))
					} else {
						params = append(params[:0], byte(r))
					}

					fn(t, params)
				}
			default:
				fn(t, nil)
			}
		}
	}
}

func (c *Cursor) reset(start, end int) {
	c.SeqType = SequenceTypeNone
	c.start = start
	c.end = end
}

func Parse(runes []rune, fn func(cursor *Cursor)) {
	state := stateStart
	cursor := &Cursor{SeqType: SequenceTypeNone, runes: &runes, start: 0, end: 0}

	for i := 0; i < len(runes); i++ {
		switch state {
		// Start state where we don't know what the sequence is, or if it even is one.
		case stateStart:
			if isCSI(runes[i:]) {
				// If we have some runes left we visit them
				if cursor.Len() > 0 {
					cursor.end = i
					fn(cursor)
				}

				cursor.start = i + 2
				cursor.end = cursor.start
				i += 1

				// We found the CSI start, so we go to the param state
				state = stateParam
			} else {
				cursor.SeqType = SequenceTypeNone
				cursor.end = i + 1
			}
		// In the param state we look for the end of the param sequence.
		case stateParam:
			if isParam(runes[i]) {
				cursor.end = i
			} else if isInter(runes[i]) {
				state = stateInter
				i -= 1
			} else if isFinal(runes[i]) {
				state = stateStart
				cursor.end = i
				cursor.SeqType = finalToType(runes[i], cursor)
				fn(cursor)
				cursor.reset(i+1, i+1)
			} else {
				state = stateStart
				cursor.reset(i, i)
			}
		// In the inter state we look for the end of the intermediate sequence.
		case stateInter:
			if isInter(runes[i]) {
				cursor.end = i
			} else if isFinal(runes[i]) {
				state = stateStart
				cursor.end = i
				cursor.SeqType = finalToType(runes[i], cursor)
				fn(cursor)
				cursor.reset(i+1, i+1)
			} else {
				state = stateStart
				cursor.reset(i, i)
			}
		}
	}

	if cursor.Len() > 0 {
		fn(cursor)
	}
}

func parseColor(rune []rune) (bool, bool, byte, byte, byte, int) {
	// Check if it can be a color sequence
	if len(rune) < 2 || !(rune[0] == '5' || rune[0] == '2') || rune[1] != ';' {
		return false, false, 0, 0, 0, 2
	}

	// 5;n case
	if rune[0] == '5' {
		// If n is not there the color id is 0
		if len(rune) == 2 {
			return true, false, 0, 0, 0, len(rune)
		}

		if num, off := parseNum(rune[2:]); off > 0 {
			return true, false, byte(num), 0, 0, 2 + off
		}

		return false, false, 0, 0, 0, 2
	}

	rgb := make([]byte, 3)
	comp := 0
	stop := 0

	// 2;r;g;b case
	for i := 2; i < len(rune); i++ {
		if v, off := parseNum(rune[i:]); off > 0 {
			i += off - 1
			rgb[comp] = byte(v)
			comp++
		}
		stop = i
		if comp == 3 {
			break
		}
	}

	return true, true, rgb[0], rgb[1], rgb[2], stop
}

func isCSI(rune []rune) bool {
	if len(rune) == 1 {
		return false
	}
	return rune[0] == csi[0] && rune[1] == csi[1]
}

func isParam(r rune) bool {
	return r >= 0x30 && r <= 0x3F
}

func isInter(r rune) bool {
	return r >= 0x20 && r <= 0x2F
}

func isFinal(r rune) bool {
	return r >= 0x40 && r <= 0x7E
}

func isNum(r rune) bool {
	return r >= 0x30 && r <= 0x39
}

func parseNum(rune []rune) (int, int) {
	if len(rune) == 0 {
		return 0, 0
	}

	var num int
	var digits int

	for _, r := range rune {
		if isNum(r) {
			num = num*10 + int(r-'0')
			digits++
		} else {
			break
		}
	}

	return num, digits
}

func finalToType(lastRune rune, cursor *Cursor) SequenceType {
	switch lastRune {
	case 'A':
		return SequenceTypeCursorUp
	case 'B':
		return SequenceTypeCursorDown
	case 'C':
		return SequenceTypeCursorForward
	case 'D':
		return SequenceTypeCursorBack
	case 'E':
		return SequenceTypeCursorNextLine
	case 'F':
		return SequenceTypeCursorPreviousLine
	case 'G':
		return SequenceTypeCursorHorizontal
	case 'H':
		return SequenceTypeCursorPosition
	case 'J':
		return SequenceTypeEraseDisplay
	case 'K':
		return SequenceTypeEraseLine
	case 'S':
		return SequenceTypeScrollUp
	case 'T':
		return SequenceTypeScrollDown
	case 's':
		return SequenceTypeSaveCursorPosition
	case 'm':
		return SequenceTypeSGR
	case 'h':
		view := cursor.View()
		if cursor.Len() >= 3 && view[len(view)-3] == '?' && view[len(view)-2] == '2' && view[len(view)-1] == '5' {
			return SequenceTypeShowCursor
		}
		return SequenceTypeUnimplemented
	case 'l':
		view := cursor.View()
		if cursor.Len() >= 3 && view[len(view)-3] == '?' && view[len(view)-2] == '2' && view[len(view)-1] == '5' {
			return SequenceTypeHideCursor
		}
		return SequenceTypeUnimplemented
	}
	return SequenceTypeNone
}
