package crt

import "image/color"

// FontWeight is the weight of a font at a certain terminal cell.
type FontWeight byte

const (
	// FontWeightNormal is the default font weight.
	FontWeightNormal FontWeight = iota

	// FontWeightBold is a bold font weight.
	FontWeightBold

	// FontWeightItalic is an italic font weight.
	FontWeightItalic
)

// GridCell is a single cell in the terminal grid.
type GridCell struct {
	Char   rune
	Fg     color.Color
	Bg     color.Color
	Weight FontWeight
}
