package main

import "math"

// PhongShader хранит данные для плоского освещения треугольника по Фонгу.
type PhongShader struct {
	lightDirection  Vec3
	cameraDirection Vec3
	cameraPosition  Vec3
	triangle        [3]Vec3
	normals         [3]Vec3
	uvs             [3]Vec2
	texture         *TGAImage
	shadow          *Renderer
	lightTransform  Mat4
}

const (
	MATERIAL_EXPONENT = 16
	shadowBias        = 0.005
)

func NewPhongShader(lightDirection, cameraDirection, cameraPosition Vec3, triangle, normals [3]Vec3, uvs [3]Vec2, texture *TGAImage, shadow *Renderer) *PhongShader {
	return &PhongShader{
		lightDirection:  lightDirection,
		cameraDirection: cameraDirection,
		cameraPosition:  cameraPosition,
		triangle:        triangle,
		normals:         normals,
		uvs:             uvs,
		texture:         texture,
		shadow:          shadow,
		lightTransform:  shadow.Perspective.Mul(shadow.ModelView),
	}
}

func (s *PhongShader) inShadow(position Vec3) bool {
	clip := s.lightTransform.MulVec(position.XYZ1())
	ndc := clip.Scale(1 / clip.W)
	screen := s.shadow.Viewport.MulVec(ndc)
	x, y := int(screen.X), int(screen.Y)
	if x < 0 || x >= s.shadow.Framebuffer.Width() || y < 0 || y >= s.shadow.Framebuffer.Height() {
		return false
	}
	depth := s.shadow.zbuffer[x+y*s.shadow.Framebuffer.Width()]
	return ndc.Z < depth-shadowBias
}

func (s *PhongShader) Fragment(bar Vec3) (discard bool, color TGAColor) {
	// world position
	frag := s.triangle[0].Scale(bar.X).
		Add(s.triangle[1].Scale(bar.Y)).
		Add(s.triangle[2].Scale(bar.Z))

	//normals
	n := s.normals[0].Scale(bar.X).Add(s.normals[1].Scale(bar.Y)).Add(s.normals[2].Scale(bar.Z)).Normalized()
	// ab := s.triangle[1].Sub(s.triangle[0])
	// ac := s.triangle[2].Sub(s.triangle[0])
	// n := ab.Cross(ac).Normalized()

	// difuse term
	cosDiffuse := n.Dot(s.lightDirection)
	diffuse := max(0, cosDiffuse)

	//speculat term
	reflect := n.Scale(cosDiffuse * 2).Sub(s.lightDirection)
	viewDirection := s.cameraPosition.Sub(frag).Normalized()
	cosSpecular := reflect.Dot(viewDirection)
	specular := math.Pow(max(0, cosSpecular), MATERIAL_EXPONENT)
	if s.inShadow(frag) {
		diffuse = 0
		specular = 0
	}

	totalLight := min(255, (specular+diffuse)*255)

	//texture
	u := s.uvs[0].X*bar.X + s.uvs[1].X*bar.Y + s.uvs[2].X*bar.Z
	v := s.uvs[0].Y*bar.X + s.uvs[1].Y*bar.Y + s.uvs[2].Y*bar.Z
	x := min(s.texture.Width()-1, max(0, int(u*float64(s.texture.Width()))))
	y := min(s.texture.Height()-1, max(0, int(v*float64(s.texture.Height()))))

	textureColor := s.texture.Get(x, y)
	lightIntensity := totalLight / 255

	finalColor := TGAColor{
		B: uint8(float64(textureColor.B) * lightIntensity),
		G: uint8(float64(textureColor.G) * lightIntensity),
		R: uint8(float64(textureColor.R) * lightIntensity),
		A: 255,
	}
	return false, finalColor
}
