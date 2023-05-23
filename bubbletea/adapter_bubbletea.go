package bubbletea

import (
	"github.com/BigJk/crt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"unicode"
)

type teaKey struct {
	key  tea.KeyType
	rune []rune
}

func repeatingKeyPressed(key ebiten.Key) bool {
	const (
		delay    = 30
		interval = 3
	)
	d := inpututil.KeyPressDuration(key)
	if d == 1 {
		return true
	}
	if d >= delay && (d-delay)%interval == 0 {
		return true
	}
	return false
}

var ebitenToTeaKeys = map[ebiten.Key]teaKey{
	ebiten.KeyEnter:      {tea.KeyEnter, []rune{'\n'}},
	ebiten.KeyTab:        {tea.KeyTab, []rune{}},
	ebiten.KeyBackspace:  {tea.KeyBackspace, []rune{}},
	ebiten.KeyDelete:     {tea.KeyDelete, []rune{}},
	ebiten.KeyHome:       {tea.KeyHome, []rune{}},
	ebiten.KeyEnd:        {tea.KeyEnd, []rune{}},
	ebiten.KeyPageUp:     {tea.KeyPgUp, []rune{}},
	ebiten.KeyArrowUp:    {tea.KeyUp, []rune{}},
	ebiten.KeyArrowDown:  {tea.KeyDown, []rune{}},
	ebiten.KeyArrowLeft:  {tea.KeyLeft, []rune{}},
	ebiten.KeyArrowRight: {tea.KeyRight, []rune{}},
	ebiten.KeyEscape:     {tea.KeyEscape, []rune{}},
	ebiten.KeyF1:         {tea.KeyF1, []rune{}},
	ebiten.KeyF2:         {tea.KeyF2, []rune{}},
	ebiten.KeyF3:         {tea.KeyF3, []rune{}},
	ebiten.KeyF4:         {tea.KeyF4, []rune{}},
	ebiten.KeyF5:         {tea.KeyF5, []rune{}},
	ebiten.KeyF6:         {tea.KeyF6, []rune{}},
	ebiten.KeyF7:         {tea.KeyF7, []rune{}},
	ebiten.KeyF8:         {tea.KeyF8, []rune{}},
	ebiten.KeyF9:         {tea.KeyF9, []rune{}},
	ebiten.KeyF10:        {tea.KeyF10, []rune{}},
	ebiten.KeyF11:        {tea.KeyF11, []rune{}},
	ebiten.KeyF12:        {tea.KeyF12, []rune{}},
	ebiten.KeyShift:      {tea.KeyShiftLeft, []rune{}},
}

var ebitenToCtrlKeys = map[ebiten.Key]tea.KeyType{
	ebiten.KeyA:            tea.KeyCtrlA,
	ebiten.KeyB:            tea.KeyCtrlB,
	ebiten.KeyC:            tea.KeyCtrlC,
	ebiten.KeyD:            tea.KeyCtrlD,
	ebiten.KeyE:            tea.KeyCtrlE,
	ebiten.KeyF:            tea.KeyCtrlF,
	ebiten.KeyG:            tea.KeyCtrlG,
	ebiten.KeyH:            tea.KeyCtrlH,
	ebiten.KeyI:            tea.KeyCtrlI,
	ebiten.KeyJ:            tea.KeyCtrlJ,
	ebiten.KeyK:            tea.KeyCtrlK,
	ebiten.KeyL:            tea.KeyCtrlL,
	ebiten.KeyM:            tea.KeyCtrlM,
	ebiten.KeyN:            tea.KeyCtrlN,
	ebiten.KeyO:            tea.KeyCtrlO,
	ebiten.KeyP:            tea.KeyCtrlP,
	ebiten.KeyQ:            tea.KeyCtrlQ,
	ebiten.KeyR:            tea.KeyCtrlR,
	ebiten.KeyS:            tea.KeyCtrlS,
	ebiten.KeyT:            tea.KeyCtrlT,
	ebiten.KeyU:            tea.KeyCtrlU,
	ebiten.KeyV:            tea.KeyCtrlV,
	ebiten.KeyW:            tea.KeyCtrlW,
	ebiten.KeyX:            tea.KeyCtrlX,
	ebiten.KeyY:            tea.KeyCtrlY,
	ebiten.KeyZ:            tea.KeyCtrlZ,
	ebiten.KeyLeftBracket:  tea.KeyCtrlOpenBracket,
	ebiten.KeyBackslash:    tea.KeyCtrlBackslash,
	ebiten.KeyRightBracket: tea.KeyCtrlCloseBracket,
	ebiten.KeyApostrophe:   tea.KeyCtrlCaret,
}

var ebitenToTeaMouse = map[ebiten.MouseButton]tea.MouseEventType{
	ebiten.MouseButtonLeft:   tea.MouseLeft,
	ebiten.MouseButtonMiddle: tea.MouseMiddle,
	ebiten.MouseButtonRight:  tea.MouseRight,
}

// Options are used to configure the adapter.
type Options func(*Adapter)

// WithFilterMousePressed filters the MousePressed event and only emits MouseReleased events.
func WithFilterMousePressed(filter bool) Options {
	return func(b *Adapter) {
		b.filterMousePressed = filter
	}
}

// Adapter represents a bubbletea adapter for the crt package.
type Adapter struct {
	prog               *tea.Program
	filterMousePressed bool
}

// NewAdapter creates a new bubbletea adapter.
func NewAdapter(prog *tea.Program, options ...Options) *Adapter {
	b := &Adapter{prog: prog, filterMousePressed: true}

	for i := range options {
		options[i](b)
	}

	return b
}

func (b *Adapter) HandleMouseMotion(motion crt.MouseMotion) {
	b.prog.Send(tea.MouseMsg{
		X:    motion.X,
		Y:    motion.Y,
		Alt:  false,
		Ctrl: false,
		Type: tea.MouseMotion,
	})
}

func (b *Adapter) HandleMouseButton(button crt.MouseButton) {
	// Filter this event or two events will be sent for one click in the current bubbletea version.
	if b.filterMousePressed && button.JustPressed {
		return
	}

	b.prog.Send(tea.MouseMsg{
		X:    button.X,
		Y:    button.Y,
		Alt:  ebiten.IsKeyPressed(ebiten.KeyAlt),
		Ctrl: ebiten.IsKeyPressed(ebiten.KeyControl),
		Type: ebitenToTeaMouse[button.Button],
	})
}

func (b *Adapter) HandleMouseWheel(wheel crt.MouseWheel) {
	if wheel.DY > 0 {
		b.prog.Send(tea.MouseMsg{
			X:    wheel.X,
			Y:    wheel.Y,
			Alt:  ebiten.IsKeyPressed(ebiten.KeyAlt),
			Ctrl: ebiten.IsKeyPressed(ebiten.KeyControl),
			Type: tea.MouseWheelUp,
		})
	} else if wheel.DY < 0 {
		b.prog.Send(tea.MouseMsg{
			X:    wheel.X,
			Y:    wheel.Y,
			Alt:  ebiten.IsKeyPressed(ebiten.KeyAlt),
			Ctrl: ebiten.IsKeyPressed(ebiten.KeyControl),
			Type: tea.MouseWheelDown,
		})
	}
}

func (b *Adapter) HandleKeyPress() {
	newInputs := ebiten.AppendInputChars([]rune{})
	for _, v := range newInputs {
		switch v {
		case ' ':
			b.prog.Send(tea.KeyMsg{
				Type:  tea.KeySpace,
				Runes: []rune{v},
				Alt:   ebiten.IsKeyPressed(ebiten.KeyAlt),
			})
		default:
			b.prog.Send(tea.KeyMsg{
				Type:  tea.KeyRunes,
				Runes: []rune{v},
				Alt:   ebiten.IsKeyPressed(ebiten.KeyAlt),
			})
		}
	}

	var keys []ebiten.Key
	keys = inpututil.AppendJustPressedKeys(keys)
	repeatedBackspace := repeatingKeyPressed(ebiten.KeyBackspace)

	if repeatedBackspace {
		b.prog.Send(tea.KeyMsg{
			Type:  tea.KeyBackspace,
			Runes: []rune{},
			Alt:   false,
		})
	}

	for _, k := range keys {
		if ebiten.IsKeyPressed(ebiten.KeyControl) {
			if tk, ok := ebitenToCtrlKeys[k]; ok {
				b.prog.Send(tea.KeyMsg{
					Type:  tk,
					Runes: []rune{},
					Alt:   false,
				})
				continue
			}
		}

		if repeatedBackspace && k == ebiten.KeyBackspace {
			continue
		}

		if val, ok := ebitenToTeaKeys[k]; ok {
			runes := make([]rune, len(val.rune))
			copy(runes, val.rune)

			if ebiten.IsKeyPressed(ebiten.KeyShift) {
				for i := range runes {
					runes[i] = unicode.ToUpper(runes[i])
				}
			}

			b.prog.Send(tea.KeyMsg{
				Type:  val.key,
				Runes: runes,
				Alt:   ebiten.IsKeyPressed(ebiten.KeyAlt),
			})
		}
	}
}

func (b *Adapter) HandleWindowSize(size crt.WindowSize) {
	b.prog.Send(tea.WindowSizeMsg{
		Width:  size.Width,
		Height: size.Height,
	})
}
