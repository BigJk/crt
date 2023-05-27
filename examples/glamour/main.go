package main

import (
	"github.com/BigJk/crt"
	bubbleadapter "github.com/BigJk/crt/bubbletea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"image/color"
	"os"
	"strings"
)

const (
	Width  = 700
	Height = 900
)

var (
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render
	readme    = ""
)

type example struct {
	viewport viewport.Model
}

func newExample() *example {
	vp := viewport.New(20, 20)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		PaddingRight(2)

	return &example{
		viewport: vp,
	}
}

func (e example) Init() tea.Cmd {
	return nil
}

func (e example) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return e, tea.Quit
		default:
			var cmd tea.Cmd
			e.viewport, cmd = e.viewport.Update(msg)
			return e, cmd
		}
	case tea.WindowSizeMsg:
		e.viewport.Width = msg.Width
		e.viewport.Height = msg.Height - 3

		renderer, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(msg.Width),
		)
		if err != nil {
			panic(err)
		}

		str, err := renderer.Render(readme)
		if err != nil {
			panic(err)
		}

		e.viewport.SetContent(str)
	}
	return e, nil
}

func (e example) View() string {
	return e.viewport.View() + e.helpView()
}

func (e example) helpView() string {
	return helpStyle("\n  ↑/↓: Navigate • q: Quit\n")
}

func main() {
	// Read the readme from repo
	f, err := os.ReadFile("./README.md")
	if err != nil {
		panic(err)
	}
	readme = string(f)
	readme = strings.Replace(readme, "\t", "    ", -1)

	fonts, err := crt.LoadFaces("./fonts/IosevkaTermNerdFontMono-Regular.ttf", "./fonts/IosevkaTermNerdFontMono-Bold.ttf", "./fonts/IosevkaTermNerdFontMono-Italic.ttf", crt.GetFontDPI(), 12.0)
	if err != nil {
		panic(err)
	}

	win, _, err := bubbleadapter.Window(Width, Height, fonts, newExample(), color.RGBA{
		R: 30,
		G: 30,
		B: 30,
		A: 255,
	}, tea.WithAltScreen())
	if err != nil {
		panic(err)
	}

	if err := win.Run("Glamour Markdown"); err != nil {
		panic(err)
	}
}
