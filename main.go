package main

import (
	"log"
	"math/rand/v2"
	"os"
)

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

	models := make([]*Model, 0, len(files))
	for _, file := range files {
		model, err := readModel(file)
		if err != nil {
			log.Fatal(err)
		}
		models = append(models, model)
	}

	texture, err := ReadTGAFile("diablo3_pose_diffuse.tga")
	if err != nil {
		log.Fatal(err)
	}

	gl := NewRenderer(width, height)
	gl.LookAt(eye, center, up)
	gl.SetPerspective(eye.Sub(center).Norm())
	gl.SetViewport(width/16, height/16, width*7/8, height*7/8)

	if err := show(gl, models, texture, cameraDirection, eye, lightDirection); err != nil {
		log.Fatal(err)
	}
}
