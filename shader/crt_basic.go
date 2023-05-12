package shader

import (
	"github.com/hajimehoshi/ebiten/v2"
	"math/rand"
)

// crtBasicKage is a CRT shader that simulates a CRT monitor with a basic pixel grid.
//
// Credits: https://quasilyte.dev/blog/post/ebitengine-shaders/
var crtBasicKage = []byte(`
	package main

	var Seed float
	var Tick float

	func tex2pixCoord(texCoord vec2) vec2 {
		pixSize := imageSrcTextureSize()
		originTexCoord, _ := imageSrcRegionOnTexture()
		actualTexCoord := texCoord - originTexCoord
		actualPixCoord := actualTexCoord * pixSize
		return actualPixCoord
	}

	func pix2texCoord(actualPixCoord vec2) vec2 {
		pixSize := imageSrcTextureSize()
		actualTexCoord := actualPixCoord / pixSize
		originTexCoord, _ := imageSrcRegionOnTexture()
		texCoord := actualTexCoord + originTexCoord
		return texCoord
	}

	func applyPixPick(pixCoord vec2, dist float, m, hash int) vec2 {
		dir := hash % m
		if dir == int(0) {
			pixCoord.x += dist
		} else if dir == int(1) {
			pixCoord.x -= dist
		} else if dir == int(2) {
			pixCoord.y += dist
		} else if dir == int(3) {
			pixCoord.y -= dist
		}
		// Otherwise, don't move it anywhere.
		return pixCoord
	}

	func shaderRand(pixCoord vec2) (seedMod, randValue int) {
		pixSize := imageSrcTextureSize()
		pixelOffset := int(pixCoord.x) + int(pixCoord.y*pixSize.x)
		seedMod = pixelOffset % int(Seed)
		pixelOffset += seedMod
		return seedMod, pixelOffset + int(Seed)
	}

	func applyVideoDegradation(y float, c vec4) vec4 {
		if c.a != 0.0 {
			// Every 4th pixel on the Y axis will be darkened.
			if int(y+Tick)%4 != int(0) {
				return c * 0.8
			}
		}
		return c
	}

	func Fragment(pos vec4, texCoord vec2, _ vec4) vec4 {
		c := imageSrc0At(texCoord)
	
		actualPixCoord := tex2pixCoord(texCoord)
		if c.a != 0.0 {
			seedMod, h := shaderRand(actualPixCoord)
			dist := 1.0
			if seedMod == int(0) {
				dist = 2.0
			}
			p := applyPixPick(actualPixCoord, dist, 10, h)
			return applyVideoDegradation(pos.y, imageSrc0At(pix2texCoord(p)))
		}
	
		return c
	}
`)

type CrtBasic struct {
	BaseShader
	tick float64
}

func (b *CrtBasic) Apply(screen *ebiten.Image, buffer *ebiten.Image) error {
	b.Lock()
	defer b.Unlock()

	b.tick += 1 / 60.0
	var options ebiten.DrawRectShaderOptions
	options.GeoM.Translate(0, 0)
	options.Images[0] = buffer
	options.Uniforms = map[string]any{
		"Seed": rand.Float64() * 10000,
		"Tick": b.tick,
	}
	screen.DrawRectShader(screen.Bounds().Dx(), screen.Bounds().Dy(), b.Shader, &options)
	return nil
}

func NewCrtBasic() (*CrtBasic, error) {
	shader, err := ebiten.NewShader([]byte(crtBasicKage))
	if err != nil {
		return nil, err
	}
	return &CrtBasic{
		BaseShader: BaseShader{
			Shader: shader,
		},
	}, nil
}
