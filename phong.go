package main

import "math"

// PhongShader хранит данные для плоского освещения треугольника по Фонгу.
type PhongShader struct {
	lightDirection  Vec3
	cameraDirection Vec3
	cameraPosition  Vec3
	triangle        [3]Vec3
}

const MATERIAL_EXPONENT = 4

func NewPhongShader(lightDirection, cameraDirection, cameraPosition Vec3, triangle [3]Vec3) *PhongShader {
	return &PhongShader{
		lightDirection:  lightDirection,
		cameraDirection: cameraDirection,
		cameraPosition:  cameraPosition,
		triangle:        triangle,
	}
}

func (s *PhongShader) Fragment(frag Vec3) (discard bool, color TGAColor) {
	// difuse term
	ab := s.triangle[1].Sub(s.triangle[0])
	ac := s.triangle[2].Sub(s.triangle[0])
	n := ab.Cross(ac).Normalized()
	cosDiffuse := n.Dot(s.lightDirection)
	diffuse := max(0, cosDiffuse)

	//speculat term
	reflect := n.Scale(cosDiffuse * 2).Sub(s.lightDirection)
	viewDirection := s.cameraPosition.Sub(frag).Normalized()
	cosSpecular := reflect.Dot(viewDirection)
	specular := math.Pow(max(0, cosSpecular), MATERIAL_EXPONENT)

	totalLight := min(255, (specular+diffuse)*255)

	return false, TGAColor{uint8(totalLight), uint8(totalLight), uint8(totalLight), 255}
}
