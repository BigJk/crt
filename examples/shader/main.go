package main

import (
	"fmt"
	"github.com/BigJk/crt"
	bubbleadapter "github.com/BigJk/crt/bubbletea"
	"github.com/BigJk/crt/shader"
	"github.com/muesli/termenv"
	"image/color"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	Width  = 1000
	Height = 600
)

var packages = []string{
	"vegeutils",
	"libgardening",
	"currykit",
	"spicerack",
	"fullenglish",
	"eggy",
	"bad-kitty",
	"chai",
	"hojicha",
	"libtacos",
	"babys-monads",
	"libpurring",
	"currywurst-devel",
	"xmodmeow",
	"licorice-utils",
	"cashew-apple",
	"rock-lobster",
	"standmixer",
	"coffee-CUPS",
	"libesszet",
	"zeichenorientierte-benutzerschnittstellen",
	"schnurrkit",
	"old-socks-devel",
	"jalapeño",
	"molasses-utils",
	"xkohlrabi",
	"party-gherkin",
	"snow-peas",
	"libyuzu",
}

func getPackages() []string {
	pkgs := packages
	copy(pkgs, packages)

	rand.Shuffle(len(pkgs), func(i, j int) {
		pkgs[i], pkgs[j] = pkgs[j], pkgs[i]
	})

	for k := range pkgs {
		pkgs[k] += fmt.Sprintf("-%d.%d.%d", rand.Intn(10), rand.Intn(10), rand.Intn(10))
	}
	return pkgs
}

type model struct {
	packages []string
	index    int
	width    int
	height   int
	spinner  spinner.Model
	progress progress.Model
	done     bool
}

var (
	currentPkgNameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#e7c6ff"))
	doneStyle           = lipgloss.NewStyle().Margin(1, 2)
	checkMark           = lipgloss.NewStyle().Foreground(lipgloss.Color("#8ac926")).SetString("✓")
)

func newModel() model {
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
		progress.WithColorProfile(termenv.TrueColor),
		progress.WithGradient("#231942", "#be95c4"),
	)
	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#669bbc"))
	return model{
		packages: getPackages(),
		spinner:  s,
		progress: p,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(downloadAndInstall(m.packages[m.index]), m.spinner.Tick)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			return m, tea.Quit
		}
	case installedPkgMsg:
		if m.index >= len(m.packages)-1 {
			// Everything's been installed. We're done!
			m.done = true
			return m, tea.Quit
		}

		// Update progress bar
		progressCmd := m.progress.SetPercent(float64(m.index) / float64(len(m.packages)-1))

		m.index++
		return m, tea.Batch(
			progressCmd,
			tea.Printf("%s %s", checkMark, m.packages[m.index]), // print success message above our program
			downloadAndInstall(m.packages[m.index]),             // download the next package
		)
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case progress.FrameMsg:
		newModel, cmd := m.progress.Update(msg)
		if newModel, ok := newModel.(progress.Model); ok {
			m.progress = newModel
		}
		return m, cmd
	}
	return m, nil
}

func (m model) View() string {
	n := len(m.packages)
	w := lipgloss.Width(fmt.Sprintf("%d", n))

	if m.done {
		return doneStyle.Render(fmt.Sprintf("Done! Installed %d packages.\n", n))
	}

	pkgCount := fmt.Sprintf(" %*d/%*d", w, m.index, w, n-1)

	spin := m.spinner.View() + " "
	prog := m.progress.View()
	cellsAvail := max(0, m.width-lipgloss.Width(spin+prog+pkgCount))

	pkgName := currentPkgNameStyle.Render(m.packages[m.index])
	info := lipgloss.NewStyle().MaxWidth(cellsAvail).Render("Installing " + pkgName)

	cellsRemaining := max(0, m.width-lipgloss.Width(spin+info+prog+pkgCount))
	gap := strings.Repeat(" ", cellsRemaining)

	return spin + info + gap + prog + pkgCount
}

type installedPkgMsg string

func downloadAndInstall(pkg string) tea.Cmd {
	// This is where you'd do i/o stuff to download and install packages. In
	// our case we're just pausing for a moment to simulate the process.
	d := time.Millisecond * time.Duration(rand.Intn(500))
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return installedPkgMsg(pkg)
	})
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func main() {
	rand.Seed(time.Now().Unix())

	fonts, err := crt.LoadFaces("./fonts/IosevkaTermNerdFontMono-Regular.ttf", "./fonts/IosevkaTermNerdFontMono-Bold.ttf", "./fonts/IosevkaTermNerdFontMono-Italic.ttf", crt.GetFontDPI(), 16.0)
	if err != nil {
		panic(err)
	}

	win, _, err := bubbleadapter.Window(Width, Height, fonts, newModel(), color.Black)
	if err != nil {
		panic(err)
	}

	lotte, err := shader.NewCrtLotte()
	if err != nil {
		panic(err)
	}

	win.SetShader(lotte)

	if err := win.Run("Simple"); err != nil {
		panic(err)
	}
}
