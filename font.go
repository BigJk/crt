package crt

import (
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"os"
)

type Fonts struct {
	Normal font.Face
	Bold   font.Face
	Italic font.Face
}

// LoadFace loads a font face from a file. The dpi and size are used to generate the font face.
//
// Example: LoadFace("./fonts/Mono-Regular.ttf", 72.0, 16.0)
func LoadFace(file string, dpi float64, size float64) (font.Face, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		panic(err)
	}

	tt, err := opentype.Parse(data)
	if err != nil {
		panic(err)
	}

	face, err := opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    size,
		DPI:     dpi,
		Hinting: font.HintingNone,
	})
	if err != nil {
		return nil, err
	}

	return face, nil
}

// LoadFaces loads a set of fonts from files. The normal, bold, and italic files
// must be provided. The dpi and size are used to generate the font faces.
func LoadFaces(normal string, bold string, italic string, dpi float64, size float64) (Fonts, error) {
	normalFace, err := LoadFace(normal, dpi, size)
	if err != nil {
		return Fonts{}, err
	}

	boldFace, err := LoadFace(bold, dpi, size)
	if err != nil {
		return Fonts{}, err
	}

	italicFace, err := LoadFace(italic, dpi, size)
	if err != nil {
		return Fonts{}, err
	}

	return Fonts{
		Normal: normalFace,
		Bold:   boldFace,
		Italic: italicFace,
	}, nil
}
