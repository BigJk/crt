package shader

import (
	"github.com/hajimehoshi/ebiten/v2"
	"sync"
)

type Shader interface {
	Apply(screen *ebiten.Image, buffer *ebiten.Image) error
}

type BaseShader struct {
	sync.Mutex
	Shader   *ebiten.Shader
	Uniforms map[string]any
}

func (b *BaseShader) Apply(screen *ebiten.Image, buffer *ebiten.Image) error {
	b.Lock()
	defer b.Unlock()

	var options ebiten.DrawRectShaderOptions
	options.GeoM.Translate(0, 0)
	options.Images[0] = buffer
	options.Uniforms = b.Uniforms
	screen.DrawRectShader(screen.Bounds().Dx(), screen.Bounds().Dy(), b.Shader, &options)
	return nil
}
