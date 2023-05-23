package main

import (
	"fmt"
	"github.com/BigJk/crt"
	bubbleadapter "github.com/BigJk/crt/bubbletea"
	tea "github.com/charmbracelet/bubbletea"
	"image/color"
)

const (
	Width  = 1000
	Height = 600
)

type model struct {
	keys []tea.KeyMsg
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.keys = append(m.keys, msg)
	}
	return m, nil
}

func (m model) View() string {
	view := ""
	for _, key := range m.keys {
		view += "  " + key.String() + fmt.Sprintf(" | %v %v", key.Runes, key.Alt) + "\n"
	}
	return view
}

func main() {
	fonts, err := crt.LoadFaces("./fonts/IosevkaTermNerdFontMono-Regular.ttf", "./fonts/IosevkaTermNerdFontMono-Bold.ttf", "./fonts/IosevkaTermNerdFontMono-Italic.ttf", crt.GetFontDPI(), 16.0)
	if err != nil {
		panic(err)
	}

	win, _, err := bubbleadapter.Window(Width, Height, fonts, model{}, color.Black)
	if err != nil {
		panic(err)
	}

	if err := win.Run("Simple"); err != nil {
		panic(err)
	}
}
