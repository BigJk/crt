package main

import (
	"github.com/BigJk/crt"
	bubbleadapter "github.com/BigJk/crt/bubbletea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"image/color"
)

const (
	Width  = 1000
	Height = 600
)

type model struct {
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m model) View() string {
	return lipgloss.NewStyle().Margin(5).Padding(5).Border(lipgloss.ThickBorder(), true).Foreground(lipgloss.Color("#ff00ff")).Render("Hello World!")
}

func main() {
	fonts, err := crt.LoadFaces("./fonts/IosevkaTermNerdFontMono-Regular.ttf", "./fonts/IosevkaTermNerdFontMono-Bold.ttf", "./fonts/IosevkaTermNerdFontMono-Italic.ttf", 72.0, 16.0)
	if err != nil {
		panic(err)
	}

	win, prog, err := bubbleadapter.Window(Width, Height, fonts, model{}, color.Black, tea.WithAltScreen())
	if err != nil {
		panic(err)
	}

	prog.Send(tea.ShowCursor())
	win.SetCursorChar("_")

	if err := win.Run("Simple"); err != nil {
		panic(err)
	}
}
