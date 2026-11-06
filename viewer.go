package main

import "github.com/hajimehoshi/ebiten/v2"

const (
	targetFPS    = 60
	lightZStep   = 0.05
	rotationStep = 0.03
)

type game struct {
	renderer        *Renderer
	models          []*Model
	texture         *TGAImage
	cameraDirection Vec3
	cameraPosition  Vec3
	lightDirection  Vec3
	lightZ          float64
	modelAngle      float64
	image           *ebiten.Image
	pixels          []byte
}

func show(renderer *Renderer, models []*Model, texture *TGAImage, cameraDirection, cameraPosition, lightDirection Vec3) error {
	width, height := renderer.Framebuffer.Width(), renderer.Framebuffer.Height()
	g := &game{
		renderer:        renderer,
		models:          models,
		texture:         texture,
		cameraDirection: cameraDirection,
		cameraPosition:  cameraPosition,
		lightDirection:  lightDirection,
		lightZ:          1,
		image:           ebiten.NewImage(width, height),
		pixels:          make([]byte, width*height*4),
	}

	ebiten.SetWindowSize(width, height)
	ebiten.SetWindowTitle("tinyrenderer")
	ebiten.SetTPS(targetFPS)
	return ebiten.RunGame(g)
}

func (g *game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		g.modelAngle -= rotationStep
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		g.modelAngle += rotationStep
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		g.lightZ += lightZStep
	}
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		g.lightZ -= lightZStep
	}

	width, height := g.renderer.Framebuffer.Width(), g.renderer.Framebuffer.Height()
	x, y := ebiten.CursorPosition()
	direction := Vec3{
		X: 2*float64(x)/float64(width) - 1,
		Y: 1 - 2*float64(y)/float64(height),
		Z: g.lightZ,
	}
	if direction.Dot(direction) > 0 {
		g.lightDirection = direction.Normalized()
	}
	g.renderFrame()
	return nil
}

func (g *game) Draw(screen *ebiten.Image) {
	framebuffer := g.renderer.Framebuffer
	width, height := framebuffer.Width(), framebuffer.Height()

	// TGAImage хранит BGR снизу вверх, а Ebitengine принимает RGBA сверху вниз.
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			src := (x + (height-1-y)*width) * int(framebuffer.format)
			dst := (x + y*width) * 4
			g.pixels[dst] = framebuffer.data[src+2]
			g.pixels[dst+1] = framebuffer.data[src+1]
			g.pixels[dst+2] = framebuffer.data[src]
			g.pixels[dst+3] = 255
		}
	}
	g.image.WritePixels(g.pixels)
	screen.DrawImage(g.image, nil)
}

func (g *game) Layout(_, _ int) (int, int) {
	return g.renderer.Framebuffer.Width(), g.renderer.Framebuffer.Height()
}

func (g *game) renderFrame() {
	gl := g.renderer
	gl.Framebuffer.Fill(NewColor(177, 195, 209, 255))
	for i := range gl.zbuffer {
		gl.zbuffer[i] = -1000
	}

	for _, model := range g.models {
		for f := 0; f < model.nfaces(); f++ {
			triangle := [3]Vec3{model.faceVert(f, 0), model.faceVert(f, 1), model.faceVert(f, 2)}
			normals := [3]Vec3{model.faceNormal(f, 0), model.faceNormal(f, 1), model.faceNormal(f, 2)}
			uvs := [3]Vec2{model.faceUV(f, 0), model.faceUV(f, 1), model.faceUV(f, 2)}
			for i := range triangle {
				triangle[i] = triangle[i].RotateY(g.modelAngle)
				normals[i] = normals[i].RotateY(g.modelAngle)
			}
			shader := NewPhongShader(g.lightDirection, g.cameraDirection, g.cameraPosition, triangle, normals, uvs, g.texture)
			var clip [3]Vec4
			for i, vertex := range triangle {
				clip[i] = gl.Perspective.MulVec(gl.ModelView.MulVec(vertex.XYZ1()))
			}
			gl.Rasterize(clip, shader)
		}
	}
}
