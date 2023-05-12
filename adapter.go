package crt

import "github.com/hajimehoshi/ebiten/v2"

type WindowSize struct {
	Width  int
	Height int
}

type MouseButton struct {
	Button       ebiten.MouseButton
	X            int
	Y            int
	Shift        bool
	Alt          bool
	Ctrl         bool
	JustPressed  bool
	JustReleased bool
}

type MouseMotion struct {
	X int
	Y int
}

type MouseWheel struct {
	X     int
	Y     int
	DX    float64
	DY    float64
	Shift bool
	Alt   bool
	Ctrl  bool
}

type KeyPress struct {
	Key   ebiten.Key
	Runes []rune
	Shift bool
	Alt   bool
	Ctrl  bool
}

type InputAdapter interface {
	HandleMouseButton(button MouseButton)
	HandleMouseMotion(motion MouseMotion)
	HandleMouseWheel(wheel MouseWheel)
	HandleKeyPress()
	HandleWindowSize(size WindowSize)
}
