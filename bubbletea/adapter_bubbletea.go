package bubbletea

import (
	"github.com/BigJk/crt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"strings"
)

var ebitenToTeaKeys = map[ebiten.Key]tea.KeyType{
	ebiten.KeyEnter:      tea.KeyEnter,
	ebiten.KeyTab:        tea.KeyTab,
	ebiten.KeySpace:      tea.KeySpace,
	ebiten.KeyBackspace:  tea.KeyBackspace,
	ebiten.KeyDelete:     tea.KeyDelete,
	ebiten.KeyHome:       tea.KeyHome,
	ebiten.KeyEnd:        tea.KeyEnd,
	ebiten.KeyPageUp:     tea.KeyPgUp,
	ebiten.KeyArrowUp:    tea.KeyUp,
	ebiten.KeyArrowDown:  tea.KeyDown,
	ebiten.KeyArrowLeft:  tea.KeyLeft,
	ebiten.KeyArrowRight: tea.KeyRight,
	ebiten.KeyEscape:     tea.KeyEscape,
}

var ebitenToTeaRunes = map[ebiten.Key][]rune{
	ebiten.Key1: {'1'},
	ebiten.Key2: {'2'},
	ebiten.Key3: {'3'},
	ebiten.Key4: {'4'},
	ebiten.Key5: {'5'},
	ebiten.Key6: {'6'},
	ebiten.Key7: {'7'},
	ebiten.Key8: {'8'},
	ebiten.Key9: {'9'},
	ebiten.Key0: {'0'},
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
	var keys []ebiten.Key
	keys = inpututil.AppendJustReleasedKeys(keys)

	for _, k := range keys {
		if val, ok := ebitenToTeaKeys[k]; ok {
			runes := []rune(strings.ToLower(k.String()))
			b.prog.Send(tea.KeyMsg{
				Type:  val,
				Runes: runes,
				Alt:   ebiten.IsKeyPressed(ebiten.KeyAlt),
			})
		} else {
			runes := []rune(strings.ToLower(k.String()))
			if val, ok := ebitenToTeaRunes[k]; ok {
				runes = val
			}
			b.prog.Send(tea.KeyMsg{
				Type:  tea.KeyRunes,
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
