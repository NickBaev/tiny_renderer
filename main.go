package main

import (
	"log"
	"math/rand/v2"
	"os"
)

const outDir = "out"

// RandomShader красит каждый треугольник своим случайным цветом.
type RandomShader struct {
	gl    *Renderer
	model *Model
	color TGAColor
}

func NewRandomShader(gl *Renderer, model *Model) *RandomShader {
	return &RandomShader{
		gl:    gl,
		model: model,
		color: NewColor(uint8(rand.IntN(256)), uint8(rand.IntN(256)), uint8(rand.IntN(256)), 255),
	}
}

// Vertex переводит вершину модели в clip-координаты.
func (s *RandomShader) Vertex(face, vert int) Vec4 {
	v := s.model.faceVert(face, vert)
	return s.gl.Perspective.MulVec(s.gl.ModelView.MulVec(v.XYZ1()))
}

func (s *RandomShader) Fragment(Vec3) (bool, TGAColor) {
	return false, s.color
}

func main() {
	const (
		width  = 800
		height = 800
	)
	var (
		eye    = Vec3{-1, 0, 2} // положение камеры
		center = Vec3{0, 0, 0}  // куда смотрит камера
		up     = Vec3{0, 1, 0}  // вектор "вверх" в мировых координатах
	)
	cameraDirection := eye.Sub(center).Normalized()
	lightDirection := Vec3{1, 1, 1}.Normalized()

	files := os.Args[1:]
	if len(files) == 0 {
		files = []string{"diablo3_pose.obj"}
	}

	gl := NewRenderer(width, height)
	gl.LookAt(eye, center, up)
	gl.SetPerspective(eye.Sub(center).Norm())
	gl.SetViewport(width/16, height/16, width*7/8, height*7/8)
	gl.Framebuffer.Fill(NewColor(177, 195, 209, 255))

	for _, file := range files {
		model, err := readModel(file)
		if err != nil {
			log.Fatal(err)
		}
		for f := 0; f < model.nfaces(); f++ {
			triangle := [3]Vec3{model.faceVert(f, 0), model.faceVert(f, 1), model.faceVert(f, 2)}
			shader := NewPhongShader(lightDirection, cameraDirection, triangle)
			var clip [3]Vec4
			for i, vertex := range triangle {
				clip[i] = gl.Perspective.MulVec(gl.ModelView.MulVec(vertex.XYZ1()))
			}
			gl.Rasterize(clip, shader)
		}
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		log.Fatal(err)
	}
	if err := gl.Framebuffer.WritePNGFile(outDir + "/framebuffer.png"); err != nil {
		log.Fatal(err)
	}
}
