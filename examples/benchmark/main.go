package main

import (
	"flag"
	"fmt"
	"github.com/BigJk/crt"
	bubbleadapter "github.com/BigJk/crt/bubbletea"
	"github.com/BigJk/crt/shader"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hajimehoshi/ebiten/v2"
	"image/color"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"time"
)

const (
	Width  = 1000
	Height = 600
)

type model struct {
	X, Y int
}

func (m *model) Init() tea.Cmd {
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m *model) View() string {
	return lipgloss.NewStyle().Margin(m.X, 0, 0, m.Y).Padding(5).Border(lipgloss.ThickBorder(), true).Background(lipgloss.Color("#fc2022")).Foreground(lipgloss.Color("#ff00ff")).Render("Hello World!")
}

func main() {
	go func() {
		fmt.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	rand.Seed(0)

	enableShader := flag.Bool("shader", false, "Enable shader")
	flag.Parse()

	fonts, err := crt.LoadFaces("./fonts/IosevkaTermNerdFontMono-Regular.ttf", "./fonts/IosevkaTermNerdFontMono-Bold.ttf", "./fonts/IosevkaTermNerdFontMono-Italic.ttf", 72.0, 9.0)
	if err != nil {
		panic(err)
	}

	mod := &model{}
	win, prog, err := bubbleadapter.Window(Width, Height, fonts, mod, color.Black, tea.WithAltScreen())
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			mod.X = rand.Intn(win.GetCellsWidth())
			mod.Y = rand.Intn(win.GetCellsHeight())
			prog.Send(time.Now())
			time.Sleep(time.Second)
		}
	}()

	var lastStart int64
	win.SetOnPreDraw(func(screen *ebiten.Image) {
		lastStart = time.Now().UnixMicro()
	})
	win.SetOnPostDraw(func(screen *ebiten.Image) {
		elapsed := time.Now().UnixMicro() - lastStart
		if (1000 / (float64(elapsed) * 0.001)) > 500 {
			return
		}

		fmt.Printf("Frame took %d micro seconds FPS=%.2f\n", elapsed, 1000/(float64(elapsed)*0.001))
	})

	if *enableShader {
		lotte, err := shader.NewCrtLotte()
		if err != nil {
			panic(err)
		}

		win.SetShader(lotte)
	}

	win.ShowTPS(true)

	if err := win.Run("Simple"); err != nil {
		panic(err)
	}
}
