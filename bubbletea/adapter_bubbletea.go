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

type BubbleTeaAdapter struct {
	prog *tea.Program
}

func NewBubbleTeaAdapter(prog *tea.Program) *BubbleTeaAdapter {
	return &BubbleTeaAdapter{prog: prog}
}

func (b *BubbleTeaAdapter) HandleMouseMotion(motion crt.MouseMotion) {
	b.prog.Send(tea.MouseMsg{
		X:    motion.X,
		Y:    motion.Y,
		Alt:  false,
		Ctrl: false,
		Type: tea.MouseMotion,
	})
}

func (b *BubbleTeaAdapter) HandleMouseButton(button crt.MouseButton) {
	b.prog.Send(tea.MouseMsg{
		X:    button.X,
		Y:    button.Y,
		Alt:  ebiten.IsKeyPressed(ebiten.KeyAlt),
		Ctrl: ebiten.IsKeyPressed(ebiten.KeyControl),
		Type: ebitenToTeaMouse[button.Button],
	})
}

func (b *BubbleTeaAdapter) HandleMouseWheel(wheel crt.MouseWheel) {
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

func (b *BubbleTeaAdapter) HandleKeyPress() {
	var keys []ebiten.Key
	keys = inpututil.AppendJustReleasedKeys(keys)

	for _, k := range keys {
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

	for k, v := range ebitenToTeaKeys {
		if inpututil.IsKeyJustReleased(k) {
			runes := []rune(strings.ToLower(k.String()))
			b.prog.Send(tea.KeyMsg{
				Type:  v,
				Runes: runes,
				Alt:   ebiten.IsKeyPressed(ebiten.KeyAlt),
			})
		}
	}
}

func (b *BubbleTeaAdapter) HandleWindowSize(size crt.WindowSize) {
	b.prog.Send(tea.WindowSizeMsg{
		Width:  size.Width,
		Height: size.Height,
	})
}
