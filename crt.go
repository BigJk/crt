package crt

import (
	"fmt"
	"github.com/BigJk/crt/shader"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/muesli/ansi"
	"image"
	"image/color"
	"io"
	"sync"
)

type Window struct {
	sync.Mutex

	// Terminal dimensions and grid.
	grid        [][]GridCell
	cellsWidth  int
	cellsHeight int
	cellWidth   int
	cellHeight  int
	cellOffsetY int

	// Input and output.
	inputAdapter InputAdapter
	tty          io.Reader

	// Terminal cursor and color states.
	cursorX    int
	cursorY    int
	mouseCellX int
	mouseCellY int
	defaultBg  color.Color
	curFg      color.Color
	curBg      color.Color
	curWeight  FontWeight

	// Other
	showTps  bool
	fonts    Fonts
	bgColors *image.RGBA
	shader   []shader.Shader
	routine  sync.Once
	tick     float64
}

// NewGame creates a new terminal game with the given dimensions and font faces.
func NewGame(width int, height int, fonts Fonts, tty io.Reader, adapter InputAdapter, defaultBg color.Color) (*Window, error) {
	if defaultBg == nil {
		defaultBg = color.Black
	}

	bounds, _, _ := fonts.Normal.GlyphBounds([]rune("█")[0])
	size := bounds.Max.Sub(bounds.Min)

	cellWidth := size.X.Round()
	cellHeight := size.Y.Round()
	cellOffsetY := -bounds.Min.Y.Round()

	cellsWidth := width / cellWidth
	cellsHeight := height / cellHeight

	grid := make([][]GridCell, cellsHeight)
	for y := 0; y < cellsHeight; y++ {
		grid[y] = make([]GridCell, cellsWidth)
		for x := 0; x < cellsWidth; x++ {
			grid[y][x] = GridCell{
				Char:   ' ',
				Fg:     color.White,
				Bg:     defaultBg,
				Weight: FontWeightNormal,
			}
		}
	}

	game := &Window{
		inputAdapter: adapter,
		cellsWidth:   cellsWidth,
		cellsHeight:  cellsHeight,
		cellWidth:    cellWidth,
		cellHeight:   cellHeight,
		cellOffsetY:  cellOffsetY,
		fonts:        fonts,
		defaultBg:    defaultBg,
		grid:         grid,
		tty:          tty,
		bgColors:     image.NewRGBA(image.Rect(0, 0, cellsWidth*cellWidth, cellsHeight*cellHeight)),
	}

	game.inputAdapter.HandleWindowSize(WindowSize{
		Width:  cellsWidth - 1,
		Height: cellsHeight,
	})

	game.ResetSGR()
	game.RecalculateBackgrounds()

	return game, nil
}

// SetShader sets a shader that is applied to the whole screen.
func (g *Window) SetShader(shader ...shader.Shader) {
	g.shader = shader
}

// ShowTPS enables or disables the TPS counter on the top left.
func (g *Window) ShowTPS(val bool) {
	g.showTps = val
}

// ResetSGR resets the SGR attributes to their default values.
func (g *Window) ResetSGR() {
	g.curFg = color.White
	g.curBg = g.defaultBg
	g.curWeight = FontWeightNormal
}

// SetBgPixels sets a chunk of background pixels in the size of the cell.
func (g *Window) SetBgPixels(x, y int, c color.Color) {
	for i := 0; i < g.cellWidth; i++ {
		for j := 0; j < g.cellHeight; j++ {
			g.bgColors.Set(x*g.cellWidth+i, y*g.cellHeight+j, c)
		}
	}
}

func (g *Window) handleCSI(csi any) {
	switch seq := csi.(type) {
	case CursorUpSeq:
		g.cursorY -= seq.Count
		if g.cursorY < 0 {
			g.cursorY = 0
		}
	case CursorDownSeq:
		g.cursorY += seq.Count
		if g.cursorY >= g.cellsHeight {
			g.cursorY = g.cellsHeight - 1
		}
	case CursorForwardSeq:
		g.cursorX += seq.Count
		if g.cursorX >= g.cellsWidth {
			g.cursorX = g.cellsWidth - 1
		}
	case CursorBackSeq:
		g.cursorX -= seq.Count
		if g.cursorX < 0 {
			g.cursorX = 0
		}
	case CursorNextLineSeq:
		g.cursorY += seq.Count
		if g.cursorY >= g.cellsHeight {
			g.cursorY = g.cellsHeight - 1
		}
		g.cursorX = 0
	case CursorPreviousLineSeq:
		g.cursorY -= seq.Count
		if g.cursorY < 0 {
			g.cursorY = 0
		}
		g.cursorX = 0
	case CursorHorizontalSeq:
		g.cursorX = seq.Count - 1
	case CursorPositionSeq:
		g.cursorX = seq.Col - 1
		g.cursorY = seq.Row - 1

		if g.cursorX < 0 {
			g.cursorX = 0
		} else if g.cursorX >= g.cellsWidth {
			g.cursorX = g.cellsWidth - 1
		}

		if g.cursorY < 0 {
			g.cursorY = 0
		} else if g.cursorY >= g.cellsHeight {
			g.cursorY = g.cellsHeight - 1
		}
	case EraseDisplaySeq:
		if seq.Type != 2 {
			return // only support 2 (erase entire display)
		}

		for i := 0; i < g.cellsWidth; i++ {
			for j := 0; j < g.cellsHeight; j++ {
				g.grid[j][i].Char = ' '
				g.grid[j][i].Fg = color.White
				g.grid[j][i].Bg = g.defaultBg
			}
		}
	case EraseLineSeq:
		switch seq.Type {
		case 0: // erase from cursor to end of line
			for i := g.cursorX; i < g.cellsWidth-g.cursorX; i++ {
				g.grid[g.cursorY][g.cursorX+i].Char = ' '
				g.grid[g.cursorY][g.cursorX+i].Fg = color.White
				g.grid[g.cursorY][g.cursorX+i].Bg = g.defaultBg
				g.SetBgPixels(g.cursorX+i, g.cursorY, g.defaultBg)
			}
		case 1: // erase from start of line to cursor
			for i := 0; i < g.cursorX; i++ {
				g.grid[g.cursorY][i].Char = ' '
				g.grid[g.cursorY][i].Fg = color.White
				g.grid[g.cursorY][i].Bg = g.defaultBg
				g.SetBgPixels(i, g.cursorY, g.defaultBg)
			}
		case 2: // erase entire line
			for i := 0; i < g.cellsWidth; i++ {
				g.grid[g.cursorY][i].Char = ' '
				g.grid[g.cursorY][i].Fg = color.White
				g.grid[g.cursorY][i].Bg = g.defaultBg
				g.SetBgPixels(i, g.cursorY, g.defaultBg)
			}
		}
	case ScrollUpSeq:
		fmt.Println("UNSUPPORTED: ScrollUpSeq", seq.Count)
	case ScrollDownSeq:
		fmt.Println("UNSUPPORTED: ScrollDownSeq", seq.Count)
	case SaveCursorPositionSeq:
		fmt.Println("UNSUPPORTED: SaveCursorPositionSeq")
	case RestoreCursorPositionSeq:
		fmt.Println("UNSUPPORTED: RestoreCursorPositionSeq")
	case ChangeScrollingRegionSeq:
		fmt.Println("UNSUPPORTED: ChangeScrollingRegionSeq")
	case InsertLineSeq:
		fmt.Println("UNSUPPORTED: InsertLineSeq")
	case DeleteLineSeq:
		fmt.Println("UNSUPPORTED: DeleteLineSeq")
	}
}

func (g *Window) handleSGR(sgr any) {
	switch seq := sgr.(type) {
	case SGRReset:
		g.ResetSGR()
	case SGRBold:
		g.curWeight = FontWeightBold
	case SGRItalic:
		g.curWeight = FontWeightItalic
	case SGRUnsetBold:
		g.curWeight = FontWeightNormal
	case SGRUnsetItalic:
		g.curWeight = FontWeightNormal
	case SGRFgTrueColor:
		g.curFg = color.RGBA{seq.R, seq.G, seq.B, 255}
	case SGRBgTrueColor:
		g.curBg = color.RGBA{seq.R, seq.G, seq.B, 255}
	}
}

func (g *Window) parseSequences(str string, printExtra bool) int {
	runes := []rune(str)

	lastFound := 0
	for i := 0; i < len(runes); i++ {
		if sgr, ok := extractSGR(string(runes[i:])); ok {
			i += len(sgr) - 1

			if sgr, ok := parseSGR(sgr); ok {
				lastFound = i
				for i := range sgr {
					g.handleSGR(sgr[i])
				}
			}
		} else if csi, ok := extractCSI(string(runes[i:])); ok {
			i += len(csi) - 1

			if csi, ok := parseCSI(csi); ok {
				lastFound = i
				g.handleCSI(csi)
			}
		} else if printExtra {
			g.PrintChar(runes[i], g.curFg, g.curBg, g.curWeight)
		}
	}
	return lastFound
}

// RecalculateBackgrounds syncs the background colors to the background pixels.
func (g *Window) RecalculateBackgrounds() {
	for i := 0; i < g.cellsWidth; i++ {
		for j := 0; j < g.cellsHeight; j++ {
			g.SetBgPixels(i, j, g.grid[j][i].Bg)
		}
	}
}

// PrintChar prints a character to the screen.
func (g *Window) PrintChar(r rune, fg, bg color.Color, weight FontWeight) {
	if r == '\n' {
		g.cursorX = 0
		g.cursorY++
		return
	}

	if ansi.PrintableRuneWidth(string(r)) == 0 {
		return
	}

	// Wrap around if we're at the end of the line.
	if g.cursorX >= g.cellsWidth {
		g.cursorX = 0
		g.cursorY++
	}

	// Scroll down if we're at the bottom and add a new line.
	if g.cursorY >= g.cellsHeight {
		diff := g.cursorY - g.cellsHeight + 1
		g.grid = g.grid[diff:]
		for i := 0; i < diff; i++ {
			g.grid = append(g.grid, make([]GridCell, g.cellsWidth))
			for i := 0; i < g.cellsWidth; i++ {
				g.grid[len(g.grid)-1][i].Char = ' '
				g.grid[len(g.grid)-1][i].Fg = color.White
				g.grid[len(g.grid)-1][i].Bg = g.defaultBg
			}
		}
		g.cursorY = g.cellsHeight - 1
		g.RecalculateBackgrounds()
	}

	// Set the cell.
	g.grid[g.cursorY][g.cursorX].Char = r
	g.grid[g.cursorY][g.cursorX].Fg = fg
	g.grid[g.cursorY][g.cursorX].Bg = bg
	g.grid[g.cursorY][g.cursorX].Weight = weight

	// Set the pixels.
	g.SetBgPixels(g.cursorX, g.cursorY, g.grid[g.cursorY][g.cursorX].Bg)

	// Move the cursor.
	g.cursorX++
}

func (g *Window) Update() error {
	g.routine.Do(func() {
		go func() {
			buf := make([]byte, 1024)
			for {
				n, err := g.tty.Read(buf)
				if err != nil {
					fmt.Println("ERROR: ", err)
					continue
				}

				if n == 0 {
					continue
				}

				g.Lock()
				{
					line := string(buf[:n])
					g.parseSequences(line, true)
				}
				g.Unlock()
			}
		}()
	})

	mx, my := ebiten.CursorPosition()
	mcx, mcy := mx/g.cellWidth, my/g.cellHeight

	if mcx != g.mouseCellX || mcy != g.mouseCellY {
		g.mouseCellX = mcx
		g.mouseCellY = mcy

		g.inputAdapter.HandleMouseMotion(MouseMotion{
			X: g.mouseCellX,
			Y: g.mouseCellY,
		})
	}

	// Mouse buttons.
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		g.inputAdapter.HandleMouseButton(MouseButton{
			X:            g.mouseCellX,
			Y:            g.mouseCellY,
			Shift:        ebiten.IsKeyPressed(ebiten.KeyShift),
			Alt:          ebiten.IsKeyPressed(ebiten.KeyAlt),
			Ctrl:         ebiten.IsKeyPressed(ebiten.KeyControl),
			Button:       ebiten.MouseButtonLeft,
			JustPressed:  false,
			JustReleased: true,
		})
	} else if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.inputAdapter.HandleMouseButton(MouseButton{
			X:            g.mouseCellX,
			Y:            g.mouseCellY,
			Shift:        ebiten.IsKeyPressed(ebiten.KeyShift),
			Alt:          ebiten.IsKeyPressed(ebiten.KeyAlt),
			Ctrl:         ebiten.IsKeyPressed(ebiten.KeyControl),
			Button:       ebiten.MouseButtonLeft,
			JustPressed:  true,
			JustReleased: false,
		})
	}

	// Mouse wheel.
	_, wy := ebiten.Wheel()
	if wy > 0 || wy < 0 {
		g.inputAdapter.HandleMouseWheel(MouseWheel{
			X:     g.mouseCellX,
			Y:     g.mouseCellY,
			Shift: ebiten.IsKeyPressed(ebiten.KeyShift),
			Alt:   ebiten.IsKeyPressed(ebiten.KeyAlt),
			Ctrl:  ebiten.IsKeyPressed(ebiten.KeyControl),
			DX:    0,
			DY:    wy,
		})
	}

	// Keyboard.
	g.inputAdapter.HandleKeyPress()

	return nil
}

func (g *Window) Draw(screen *ebiten.Image) {
	g.Lock()
	defer g.Unlock()

	screen.Fill(g.defaultBg)

	bufferImage := ebiten.NewImage(g.cellsWidth*g.cellWidth, g.cellsHeight*g.cellHeight)

	// Draw background
	bufferImage.WritePixels(g.bgColors.Pix)

	// Draw text
	for y := 0; y < g.cellsHeight; y++ {
		for x := 0; x < g.cellsWidth; x++ {
			if g.grid[y][x].Char == ' ' {
				continue
			}

			switch g.grid[y][x].Weight {
			case FontWeightNormal:
				text.Draw(bufferImage, string(g.grid[y][x].Char), g.fonts.Normal, x*g.cellWidth, y*g.cellHeight+g.cellOffsetY, g.grid[y][x].Fg)
			case FontWeightBold:
				text.Draw(bufferImage, string(g.grid[y][x].Char), g.fonts.Bold, x*g.cellWidth, y*g.cellHeight+g.cellOffsetY, g.grid[y][x].Fg)
			case FontWeightItalic:
				text.Draw(bufferImage, string(g.grid[y][x].Char), g.fonts.Italic, x*g.cellWidth, y*g.cellHeight+g.cellOffsetY, g.grid[y][x].Fg)
			}
		}
	}

	g.tick += 1 / 60.0
	if g.shader != nil {
		for i := range g.shader {
			_ = g.shader[i].Apply(screen, bufferImage)

			if len(g.shader) > 0 {
				bufferImage.DrawImage(screen, nil)
			}
		}
	} else {
		screen.DrawImage(bufferImage, nil)
	}

	if g.showTps {
		ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %0.2f", ebiten.CurrentTPS()))
	}
}

func (g *Window) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.cellsWidth * g.cellWidth, g.cellsHeight * g.cellHeight
}

func (g *Window) Run(title string) error {
	sw, sh := g.Layout(0, 0)

	ebiten.SetScreenFilterEnabled(false)
	ebiten.SetWindowSize(sw, sh)
	ebiten.SetWindowTitle(title)
	if err := ebiten.RunGame(g); err != nil {
		return err
	}

	return nil
}

func (g *Window) Kill() {
	SysKill()
}
