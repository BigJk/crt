package ansi

import (
	"bytes"
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"strings"
	"testing"
)

func TestParseNum(t *testing.T) {
	for i := 0; i < 1000; i++ {
		num := i
		if i > 255 {
			num = rand.Intn(1000)
		}
		s := fmt.Sprintf("%d;", num)
		pnum, off := parseNum([]rune(s))
		assert.Equal(t, num, pnum)
		assert.Equal(t, len(s)-1, off)
	}
}

func randomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}

func removeAnsiReset(s string) string {
	return strings.Replace(s, "\x1b[0m", "", 1)
}

func TestParse(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)

	type SGRArgs struct {
		Bold      bool
		Italic    bool
		Underline bool
		Arg1      byte
		Arg2      byte
		Arg3      byte
	}

	type Args struct {
		Arg1 byte
		Arg2 byte
		SGR  SGRArgs
	}

	type SequenceTester struct {
		Args  int
		Gen   func() (string, Args)
		Check func(cursor *Cursor, args Args)
	}

	type SelectedTester struct {
		Name string
		Args Args
	}

	sequenceTests := map[string]SequenceTester{
		"CursorUp": {
			Args: 1,
			Gen: func() (string, Args) {
				arg := byte(rand.Intn(255))
				return fmt.Sprintf(termenv.CSI+termenv.CursorUpSeq, arg), Args{
					Arg1: arg,
					Arg2: 0,
				}
			},
			Check: func(cursor *Cursor, args Args) {
				assert.Equal(t, SequenceTypeCursorUp, cursor.SeqType, "CursorUp")
				assert.Equal(t, args.Arg1, cursor.Arg1(), "CursorUp")
			},
		},
		"CursorDown": {
			Args: 1,
			Gen: func() (string, Args) {
				arg := byte(rand.Intn(255))
				return fmt.Sprintf(termenv.CSI+termenv.CursorDownSeq, arg), Args{
					Arg1: arg,
					Arg2: 0,
				}
			},
			Check: func(cursor *Cursor, args Args) {
				assert.Equal(t, SequenceTypeCursorDown, cursor.SeqType, "CursorDown")
				assert.Equal(t, args.Arg1, cursor.Arg1(), "CursorDown")
			},
		},
		"CursorForward": {
			Args: 1,
			Gen: func() (string, Args) {
				arg := byte(rand.Intn(255))
				return fmt.Sprintf(termenv.CSI+termenv.CursorForwardSeq, arg), Args{
					Arg1: arg,
					Arg2: 0,
				}
			},
			Check: func(cursor *Cursor, args Args) {
				assert.Equal(t, SequenceTypeCursorForward, cursor.SeqType, "CursorForward")
				assert.Equal(t, args.Arg1, cursor.Arg1(), "CursorForward")
			},
		},
		"CursorBack": {
			Args: 1,
			Gen: func() (string, Args) {
				arg := byte(rand.Intn(255))
				return fmt.Sprintf(termenv.CSI+termenv.CursorBackSeq, arg), Args{
					Arg1: arg,
					Arg2: 0,
				}
			},
			Check: func(cursor *Cursor, args Args) {
				assert.Equal(t, SequenceTypeCursorBack, cursor.SeqType, "CursorBack")
				assert.Equal(t, args.Arg1, cursor.Arg1(), "CursorBack")
			},
		},
		"CursorNextLine": {
			Args: 1,
			Gen: func() (string, Args) {
				arg := byte(rand.Intn(255))
				return fmt.Sprintf(termenv.CSI+termenv.CursorNextLineSeq, arg), Args{
					Arg1: arg,
					Arg2: 0,
				}
			},
			Check: func(cursor *Cursor, args Args) {
				assert.Equal(t, SequenceTypeCursorNextLine, cursor.SeqType, "CursorNextLine")
				assert.Equal(t, args.Arg1, cursor.Arg1(), "CursorNextLine")
			},
		},
		"CursorPreviousLine": {
			Args: 1,
			Gen: func() (string, Args) {
				arg := byte(rand.Intn(255))
				return fmt.Sprintf(termenv.CSI+termenv.CursorPreviousLineSeq, arg), Args{
					Arg1: arg,
					Arg2: 0,
				}
			},
			Check: func(cursor *Cursor, args Args) {
				assert.Equal(t, SequenceTypeCursorPreviousLine, cursor.SeqType, "CursorPreviousLine")
				assert.Equal(t, args.Arg1, cursor.Arg1(), "CursorPreviousLine")
			},
		},
		"CursorHorizontalAbsolute": {
			Args: 1,
			Gen: func() (string, Args) {
				arg := byte(rand.Intn(255))
				return fmt.Sprintf(termenv.CSI+termenv.CursorHorizontalSeq, arg), Args{
					Arg1: arg,
					Arg2: 0,
				}
			},
			Check: func(cursor *Cursor, args Args) {
				assert.Equal(t, SequenceTypeCursorHorizontal, cursor.SeqType, "CursorHorizontalAbsolute")
				assert.Equal(t, args.Arg1, cursor.Arg1(), "CursorHorizontalAbsolute")
			},
		},
		"CursorPosition": {
			Args: 2,
			Gen: func() (string, Args) {
				arg1 := byte(rand.Intn(255))
				arg2 := byte(rand.Intn(255))
				return fmt.Sprintf(termenv.CSI+termenv.CursorPositionSeq, arg1, arg2), Args{
					Arg1: arg1,
					Arg2: arg2,
				}
			},
			Check: func(cursor *Cursor, args Args) {
				assert.Equal(t, SequenceTypeCursorPosition, cursor.SeqType, "CursorPosition")
				arg1, arg2 := cursor.Arg2()
				assert.Equal(t, args.Arg1, arg1, "CursorPosition")
				assert.Equal(t, args.Arg2, arg2, "CursorPosition")
			},
		},
		"CursorShow": {
			Args: 0,
			Gen: func() (string, Args) {
				return fmt.Sprintf(termenv.CSI + termenv.ShowCursorSeq), Args{
					Arg1: 0,
					Arg2: 0,
				}
			},
			Check: func(cursor *Cursor, args Args) {
				assert.Equal(t, SequenceTypeShowCursor, cursor.SeqType, "CursorShow")
			},
		},
		"CursorHide": {
			Args: 0,
			Gen: func() (string, Args) {
				return fmt.Sprintf(termenv.CSI + termenv.HideCursorSeq), Args{
					Arg1: 0,
					Arg2: 0,
				}
			},
			Check: func(cursor *Cursor, args Args) {
				assert.Equal(t, SequenceTypeHideCursor, cursor.SeqType, "CursorHide")
			},
		},
		"EraseDisplay": {
			Args: 1,
			Gen: func() (string, Args) {
				arg := byte(rand.Intn(3))
				return fmt.Sprintf(termenv.CSI+termenv.EraseDisplaySeq, arg), Args{
					Arg1: arg,
					Arg2: 0,
				}
			},
			Check: func(cursor *Cursor, args Args) {
				assert.Equal(t, SequenceTypeEraseDisplay, cursor.SeqType, "EraseDisplay")
				assert.Equal(t, args.Arg1, cursor.Arg1(), "EraseDisplay")
			},
		},
		"EraseLine": {
			Args: 1,
			Gen: func() (string, Args) {
				arg := byte(rand.Intn(3))
				return fmt.Sprintf(termenv.CSI+termenv.EraseLineSeq, arg), Args{
					Arg1: arg,
					Arg2: 0,
				}
			},
			Check: func(cursor *Cursor, args Args) {
				assert.Equal(t, SequenceTypeEraseLine, cursor.SeqType, "EraseLine")
				assert.Equal(t, args.Arg1, cursor.Arg1(), "EraseLine")
			},
		},
		"ScrollUp": {
			Args: 1,
			Gen: func() (string, Args) {
				arg := byte(rand.Intn(255))
				return fmt.Sprintf(termenv.CSI+termenv.ScrollUpSeq, arg), Args{
					Arg1: arg,
					Arg2: 0,
				}
			},
			Check: func(cursor *Cursor, args Args) {
				assert.Equal(t, SequenceTypeScrollUp, cursor.SeqType, "ScrollUp")
				assert.Equal(t, args.Arg1, cursor.Arg1(), "ScrollUp")
			},
		},
		"ScrollDown": {
			Args: 1,
			Gen: func() (string, Args) {
				arg := byte(rand.Intn(255))
				return fmt.Sprintf(termenv.CSI+termenv.ScrollDownSeq, arg), Args{
					Arg1: arg,
					Arg2: 0,
				}
			},
			Check: func(cursor *Cursor, args Args) {
				assert.Equal(t, SequenceTypeScrollDown, cursor.SeqType, "ScrollDown")
			},
		},
		"SGR": {
			Args: 0,
			Gen: func() (string, Args) {
				bold := rand.Intn(2) == 1
				italic := rand.Intn(2) == 1
				underline := rand.Intn(2) == 1

				if rand.Intn(2) == 1 {
					color := byte(rand.Intn(255))
					return removeAnsiReset(lipgloss.NewStyle().Bold(bold).Italic(italic).Underline(underline).Foreground(lipgloss.Color(fmt.Sprint(color))).Render("X")), Args{
						Arg1: 0,
						Arg2: 0,
						SGR: SGRArgs{
							Bold:      bold,
							Italic:    italic,
							Underline: underline,
							Arg1:      color,
						},
					}
				}

				r := byte(rand.Intn(255))
				g := byte(rand.Intn(255))
				b := byte(rand.Intn(255))

				return removeAnsiReset(lipgloss.NewStyle().Bold(bold).Italic(italic).Underline(underline).Foreground(lipgloss.Color(fmt.Sprintf("#%02X%02X%02X", r, g, b))).Render("X")), Args{
					Arg1: 0,
					Arg2: 0,
					SGR: SGRArgs{
						Bold:      bold,
						Italic:    italic,
						Underline: underline,
						Arg1:      r,
						Arg2:      g,
						Arg3:      b,
					},
				}
			},
			Check: func(cursor *Cursor, args Args) {
				if !assert.Equal(t, SequenceTypeSGR, cursor.SeqType, "SGR") {
					return
				}

				cursor.VisitSGR(func(code SGRType, params []byte) {
					switch code {
					case SGRTypeBold:
						assert.Equal(t, true, args.SGR.Bold, "SGR")
					case SGRTypeItalic:
						assert.Equal(t, true, args.SGR.Italic, "SGR")
					case SGRTypeUnderline:
						assert.Equal(t, true, args.SGR.Underline, "SGR")
					case SGRTypeFgColor:
						if len(params) == 1 {
							assert.Equal(t, args.SGR.Arg1, params[0], "SGR")
						} else {
							// Termenv seems to change colors slightly
							assert.InDelta(t, args.SGR.Arg1, params[0], 2, "SGR")
							assert.InDelta(t, args.SGR.Arg2, params[1], 2, "SGR")
							assert.InDelta(t, args.SGR.Arg3, params[2], 2, "SGR")
						}
					}
				})
			},
		},
	}

	var possibleTests []string
	var selected []SelectedTester
	var testString string

	// Try everything at least once
	for name, test := range sequenceTests {
		possibleTests = append(possibleTests, name)

		str, expectedArgs := test.Gen()

		testString += str
		selected = append(selected, SelectedTester{
			Name: name,
			Args: expectedArgs,
		})
	}

	// Randomly generate 100 extra ones
	for i := 0; i < 1000; i++ {
		selectedName := possibleTests[rand.Intn(len(possibleTests))]
		test := sequenceTests[selectedName]
		str, expectedArgs := test.Gen()

		testString += str
		selected = append(selected, SelectedTester{
			Name: selectedName,
			Args: expectedArgs,
		})

		if rand.Intn(2) == 0 {
			testString += randomString(1 + rand.Intn(25))
		}
	}

	var i int
	Parse([]rune(testString), func(cursor *Cursor) {
		// If its a non-sequence, skip it
		if cursor.SeqType == SequenceTypeNone {
			return
		}

		sequenceTests[selected[i].Name].Check(cursor, selected[i].Args)
		i++
	})
}

func TestParseColor(t *testing.T) {
	testVal1 := []rune("2;255;0;255")
	ok, trueColor, r, g, b, off := parseColor(testVal1)
	assert.Equal(t, true, ok)
	assert.Equal(t, true, trueColor)
	assert.Equal(t, byte(255), r)
	assert.Equal(t, byte(0), g)
	assert.Equal(t, byte(255), b)
	assert.Equal(t, len(testVal1), off)

	testVal2 := []rune("5;23")
	ok, trueColor, r, g, b, off = parseColor(testVal2)
	assert.Equal(t, true, ok)
	assert.Equal(t, false, trueColor)
	assert.Equal(t, byte(23), r)
	assert.Equal(t, byte(0), g)
	assert.Equal(t, byte(0), b)
	assert.Equal(t, len(testVal2), off)

	testVal3 := []rune("5;")
	ok, trueColor, r, g, b, off = parseColor(testVal3)
	assert.Equal(t, true, ok)
	assert.Equal(t, false, trueColor)
	assert.Equal(t, byte(0), r)
	assert.Equal(t, byte(0), g)
	assert.Equal(t, byte(0), b)
	assert.Equal(t, len(testVal3), off)

	testVal4 := []rune("2;1;;1")
	ok, trueColor, r, g, b, off = parseColor(testVal4)
	assert.Equal(t, true, ok)
	assert.Equal(t, true, trueColor)
	assert.Equal(t, byte(1), r)
	assert.Equal(t, byte(0), g)
	assert.Equal(t, byte(1), b)
	assert.Equal(t, len(testVal4), off)
}

func BenchmarkParse(b *testing.B) {
	var testString string

	testString += "AbcDefG"
	testString += fmt.Sprintf(termenv.CSI+termenv.EraseDisplaySeq, 20)
	testString += "HELLO WORLD"
	testString += fmt.Sprintf(termenv.CSI+termenv.CursorPositionSeq, 1, 29)
	testString += fmt.Sprintf(termenv.CSI+termenv.CursorPositionSeq, 1, 2)
	testString += "HELLO WORLD"
	testString += termenv.CSI + termenv.ShowCursorSeq
	testString += fmt.Sprintf(termenv.CSI+termenv.CursorPositionSeq, 1, 29)
	testString += fmt.Sprintf(termenv.CSI+termenv.CursorBackSeq, 5)

	buf := &bytes.Buffer{}
	lip := lipgloss.NewRenderer(buf, termenv.WithProfile(termenv.TrueColor))
	testString += lip.NewStyle().Bold(true).Foreground(lipgloss.Color("#ff00ff")).Render("Hello World") + "asdasdasdasdasd" + lip.NewStyle().Bold(true).Foreground(lipgloss.Color("10")).Render("Hello World") + lip.NewStyle().Italic(true).Background(lipgloss.Color("#ff00ff")).Render("Hello World")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Parse([]rune(testString), func(cursor *Cursor) {})
	}
}
