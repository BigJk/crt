# crt — *cathode-ray tube*

![Screenshot](./.github/screenshot.png)

## About

CRT is a library to provide a simple terminal emulator that can be attached to a ``tea.Program``. It uses ``ebitengine`` to render the terminal. It supports TrueColor, Mouse and Keyboard input. It interprets the CSI escape sequences coming from bubbletea and renders them to the terminal.

This started as a simple proof of concept for the game I'm writing with the help of bubbletea, aclled [End Of Eden](github.com/BigJk/end_of_eden). I wanted to give people who have no clue about the terminal a simple option to play the game without interacting with the terminal directly. It's also possible to apply shaders to the terminal to give it a more retro look.

## Usage

```
go get github.com/BigJk/crt@latest
```


```go
func main() {
	fonts, err := crt.LoadFaces("./fonts/SomeFont-Regular.ttf", "./fonts/SomeFont-Bold.ttf", "./fonts/SomeFont-Italic.ttf", 72.0, 16.0)
	if err != nil {
		panic(err)
	}

	// Just pass your tea.Model to the bubbleadapter and it will render it to the terminal
	win, err := bubbleadapter.Window(Width, Height, fonts, model{}, color.Black, tea.WithAltScreen())
	if err != nil {
		panic(err)
	}

	// Star the terminal with the given title
	if err := win.Run("Simple"); err != nil {
		panic(err)
	}
}
```

## Limitations

- Only supports TrueColor at the moment (no 256 color support) so you need to use TrueColor collors in lipgloss (e.g. ``lipgloss.Color("#ff0000")``)