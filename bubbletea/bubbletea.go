package bubbletea

import (
	"fmt"
	"github.com/BigJk/crt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"image/color"
)

func init() {
	lipgloss.SetColorProfile(termenv.TrueColor)
}

type fakeEnviron struct{}

func (f fakeEnviron) Environ() []string {
	return []string{"TERM", "COLORTERM"}
}

func (f fakeEnviron) Getenv(s string) string {
	switch s {
	case "TERM":
		return "xterm-256color"
	case "COLORTERM":
		return "truecolor"
	}
	return ""
}

// Window creates a new crt based bubbletea window with the given width, height, fonts, model and default background color.
// Additional options can be passed to the bubbletea program.
func Window(width int, height int, fonts crt.Fonts, model tea.Model, defaultBg color.Color, options ...tea.ProgramOption) (*crt.Window, error) {
	gameInput := crt.NewConcurrentRW()
	gameOutput := crt.NewConcurrentRW()

	go gameInput.Run()
	go gameOutput.Run()

	prog := tea.NewProgram(
		model,
		append([]tea.ProgramOption{
			tea.WithMouseAllMotion(),
			tea.WithInput(gameInput),
			tea.WithOutput(termenv.NewOutput(gameOutput, termenv.WithEnvironment(fakeEnviron{}), termenv.WithTTY(false), termenv.WithProfile(termenv.TrueColor), termenv.WithColorCache(true))),
			tea.WithANSICompressor(),
		}, options...)...,
	)

	go func() {
		fmt.Println("Running Bubbletea program")
		if _, err := prog.Run(); err != nil {
			fmt.Printf("Alas, there's been an error: %v", err)
		}

		crt.SysKill()
	}()

	return crt.NewGame(width, height, fonts, gameOutput, NewAdapter(prog), defaultBg)
}
