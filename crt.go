package crt

import (
	"fmt"
	"github.com/BigJk/crt/shader"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/muesli/ansi"
	"github.com/muesli/termenv"
	"image"
	"image/color"
	"io"
	"sync"
	"unicode/utf8"
)

// colorCache is the ansi color cache.
var colorCache = map[int]color.Color{}

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
	cursorChar  string
	cursorColor color.Color
	showCursor  bool
	cursorX     int
	cursorY     int
	mouseCellX  int
	mouseCellY  int
	defaultBg   color.Color
	curFg       color.Color
	curBg       color.Color
	curWeight   FontWeight

	// Callbacks
	onUpdate   func()
	onPreDraw  func(screen *ebiten.Image)
	onPostDraw func(screen *ebiten.Image)

	// Other
	seqBuffer        []byte
	showTps          bool
	fonts            Fonts
	bgColors         *image.RGBA
	shader           []shader.Shader
	routine          sync.Once
	shaderByteBuffer []byte
	shaderBuffer     *ebiten.Image
	lastBuffer       *ebiten.Image
	invalidateBuffer bool
}

// NewGame creates a new terminal game with the given dimensions and font faces.
func NewGame(width int, height int, fonts Fonts, tty io.Reader, adapter InputAdapter, defaultBg color.Color) (*Window, error) {
	if defaultBg == nil {
		defaultBg = color.Black
	}

	bounds, _, _ := fonts.Normal.GlyphBounds([]rune("█")[0])
	size := bounds.Max.Sub(bounds.Min)

	cellWidth := size.X.Ceil()
	cellHeight := size.Y.Ceil()
	cellOffsetY := -bounds.Min.Y.Ceil()

	cellsWidth := int(float64(width)*DeviceScale()) / cellWidth
	cellsHeight := int(float64(height)*DeviceScale()) / cellHeight

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
		inputAdapter:     adapter,
		cellsWidth:       cellsWidth,
		cellsHeight:      cellsHeight,
		cellWidth:        cellWidth,
		cellHeight:       cellHeight,
		cellOffsetY:      cellOffsetY,
		fonts:            fonts,
		defaultBg:        defaultBg,
		grid:             grid,
		tty:              tty,
		bgColors:         image.NewRGBA(image.Rect(0, 0, cellsWidth*cellWidth, cellsHeight*cellHeight)),
		lastBuffer:       ebiten.NewImage(cellsWidth*cellWidth, cellsHeight*cellHeight),
		cursorChar:       "█",
		cursorColor:      color.RGBA{R: 255, G: 255, B: 255, A: 100},
		onUpdate:         func() {},
		onPreDraw:        func(screen *ebiten.Image) {},
		onPostDraw:       func(screen *ebiten.Image) {},
		invalidateBuffer: true,
		seqBuffer:        make([]byte, 0, 2^12),
	}

	game.inputAdapter.HandleWindowSize(WindowSize{
		Width:  cellsWidth - 1,
		Height: cellsHeight,
	})

	game.ResetSGR()
	game.RecalculateBackgrounds()

	return game, nil
}

// SetShowCursor enables or disables the cursor.
func (g *Window) SetShowCursor(val bool) {
	g.showCursor = val
	g.InvalidateBuffer()
}

// SetCursorChar sets the character that is used for the cursor.
func (g *Window) SetCursorChar(char string) {
	g.cursorChar = char
	g.InvalidateBuffer()
}

// SetCursorColor sets the color of the cursor.
func (g *Window) SetCursorColor(color color.Color) {
	g.cursorColor = color
	g.InvalidateBuffer()
}

// SetShader sets a shader that is applied to the whole screen.
func (g *Window) SetShader(shader ...shader.Shader) {
	g.shader = shader
}

// SetOnUpdate sets a function that is called every frame.
func (g *Window) SetOnUpdate(fn func()) {
	g.onUpdate = fn
}

// SetOnPreDraw sets a function that is called before the screen is drawn.
func (g *Window) SetOnPreDraw(fn func(screen *ebiten.Image)) {
	g.onPreDraw = fn
}

// SetOnPostDraw sets a function that is called after the screen is drawn.
func (g *Window) SetOnPostDraw(fn func(screen *ebiten.Image)) {
	g.onPostDraw = fn
}

// ShowTPS enables or disables the TPS counter on the top left.
func (g *Window) ShowTPS(val bool) {
	g.showTps = val
}

// InvalidateBuffer forces the buffer to be redrawn.
func (g *Window) InvalidateBuffer() {
	g.invalidateBuffer = true
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
	g.InvalidateBuffer()
}

// SetBg sets the background color of a cell and checks if it needs to be redrawn.
func (g *Window) SetBg(x, y int, c color.Color) {
	ra, rg, rb, _ := g.grid[y][x].Bg.RGBA()
	ca, cg, cb, _ := c.RGBA()
	if ra == ca && rg == cg && rb == cb {
		return
	}

	g.SetBgPixels(x, y, c)
	g.grid[y][x].Bg = c
}

// GetCellsWidth returns the number of cells in the x direction.
func (g *Window) GetCellsWidth() int {
	return g.cellsWidth
}

// GetCellsHeight returns the number of cells in the y direction.
func (g *Window) GetCellsHeight() int {
	return g.cellsHeight
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
				g.SetBg(g.cursorX+i, g.cursorY, g.defaultBg)
			}
		case 1: // erase from start of line to cursor
			for i := 0; i < g.cursorX; i++ {
				g.grid[g.cursorY][i].Char = ' '
				g.grid[g.cursorY][i].Fg = color.White
				g.SetBg(i, g.cursorY, g.defaultBg)
			}
		case 2: // erase entire line
			for i := 0; i < g.cellsWidth; i++ {
				g.grid[g.cursorY][i].Char = ' '
				g.grid[g.cursorY][i].Fg = color.White
				g.SetBg(i, g.cursorY, g.defaultBg)
			}
		}
	case CursorShowSeq:
		g.SetShowCursor(true)
	case CursorHideSeq:
		g.SetShowCursor(false)
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
		g.curFg = color.RGBA{R: seq.R, G: seq.G, B: seq.B, A: 255}
	case SGRBgTrueColor:
		g.curBg = color.RGBA{R: seq.R, G: seq.G, B: seq.B, A: 255}
	case SGRFgColor:
		if val, ok := colorCache[seq.Id]; ok {
			g.curFg = val
		} else {
			if col, err := colorful.Hex(termenv.ANSI256Color(seq.Id).String()); err == nil {
				g.curFg = col
				colorCache[seq.Id] = col
			}
		}
	case SGRBgColor:
		if val, ok := colorCache[seq.Id]; ok {
			g.curBg = val
		} else {
			if col, err := colorful.Hex(termenv.ANSI256Color(seq.Id).String()); err == nil {
				g.curBg = col
				colorCache[seq.Id] = col
			}
		}
	}
}

func (g *Window) parseSequences(str string, printExtra bool) int {
	lastFound := 0
	for i := 0; i < len(str); i++ {
		if sgr, ok := extractSGR(str[i:]); ok {
			i += len(sgr) - 1

			if sgr, ok := parseSGR(sgr); ok {
				lastFound = i
				for i := range sgr {
					g.handleSGR(sgr[i])
					g.InvalidateBuffer()
				}
			}
		} else if csi, ok := extractCSI(str[i:]); ok {
			i += len(csi) - 1

			if csi, ok := parseCSI(csi); ok {
				lastFound = i
				g.handleCSI(csi)
				g.InvalidateBuffer()
			}
		} else if printExtra {
			if r, size := utf8.DecodeRuneInString(str[i:]); r != utf8.RuneError {
				g.PrintChar(r, g.curFg, g.curBg, g.curWeight)
				i += size - 1
			}
		}
	}

	return lastFound
}

func (g *Window) drainSequence() {
	if len(g.seqBuffer) > 0 {
		g.parseSequences(string(g.seqBuffer), true)
		g.seqBuffer = g.seqBuffer[:0]
	}
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

	g.InvalidateBuffer()
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
					g.seqBuffer = append(g.seqBuffer, buf[:n]...)
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

	g.onUpdate()

	return nil
}

func (g *Window) Draw(screen *ebiten.Image) {
	g.Lock()
	defer g.Unlock()

	g.onPreDraw(screen)

	// We process the sequence buffer here so that we don't get flickering
	g.drainSequence()

	screen.Fill(g.defaultBg)

	// Get current buffer
	bufferImage := g.lastBuffer

	// Only draw the buffer if it's invalid
	if g.invalidateBuffer {
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

		// Draw cursor
		if g.showCursor {
			text.Draw(bufferImage, g.cursorChar, g.fonts.Normal, g.cursorX*g.cellWidth, g.cursorY*g.cellHeight+g.cellOffsetY, g.cursorColor)
		}

		g.lastBuffer = bufferImage
		g.invalidateBuffer = false
	}

	// Draw shader
	if g.shader != nil {
		if g.shaderBuffer == nil {
			g.shaderBuffer = ebiten.NewImageFromImage(bufferImage)
		} else {
			bounds := g.shaderBuffer.Bounds()
			if len(g.shaderByteBuffer) < 4*bounds.Dx()*bounds.Dy() {
				g.shaderByteBuffer = make([]byte, 4*bounds.Dx()*bounds.Dy())
			}
			bufferImage.ReadPixels(g.shaderByteBuffer)
			g.shaderBuffer.WritePixels(g.shaderByteBuffer)
		}

		for i := range g.shader {
			_ = g.shader[i].Apply(screen, g.shaderBuffer)

			if len(g.shader) > 0 {
				g.shaderBuffer.DrawImage(screen, nil)
			}
		}
	} else {
		screen.DrawImage(bufferImage, nil)
	}

	if g.showTps {
		ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %0.2f", ebiten.CurrentTPS()))
	}

	g.onPostDraw(screen)
}

func (g *Window) Layout(outsideWidth, outsideHeight int) (int, int) {
	s := DeviceScale()
	return int(float64(outsideWidth) * s), int(float64(outsideHeight) * s)
}

func (g *Window) Run(title string) error {
	ebiten.SetScreenFilterEnabled(false)
	ebiten.SetWindowSize(int(float64(g.cellsWidth*g.cellWidth)/DeviceScale()), int(float64(g.cellsHeight*g.cellHeight)/DeviceScale()))
	ebiten.SetWindowTitle(title)
	if err := ebiten.RunGame(g); err != nil {
		return err
	}

	return nil
}

func (g *Window) Kill() {
	SysKill()
}
