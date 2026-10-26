package main

import (
	"math"
	"runtime"
	"sync"
)

// Renderer хранит состояние конвейера: матрицы и буферы кадра и глубины.
type Renderer struct {
	ModelView, Viewport, Perspective Mat4
	Framebuffer                      *TGAImage
	zbuffer                          []float64
}

// NewRenderer создаёт RGB-буфер кадра width×height и буфер глубины того же размера.
func NewRenderer(width, height int) *Renderer {
	r := &Renderer{
		Framebuffer: NewTGAImage(width, height, RGB),
		zbuffer:     make([]float64, width*height),
	}
	for i := range r.zbuffer {
		r.zbuffer[i] = -1000
	}
	return r
}

// IShader — фрагментный шейдер. Fragment получает барицентрические координаты
// пикселя и может отбросить фрагмент. Вызывается из нескольких горутин сразу,
// поэтому не должен менять состояние шейдера.
type IShader interface {
	Fragment(bar Vec3) (discard bool, color TGAColor)
}

// LookAt строит ModelView: переносит center в начало координат и поворачивает
// мир так, чтобы камера в eye смотрела вдоль -z, а up был направлен вверх.
func (r *Renderer) LookAt(eye, center, up Vec3) {
	n := eye.Sub(center).Normalized()
	l := up.Cross(n).Normalized()
	m := n.Cross(l).Normalized()
	r.ModelView = Mat4{
		{l.X, l.Y, l.Z, 0},
		{m.X, m.Y, m.Z, 0},
		{n.X, n.Y, n.Z, 0},
		{0, 0, 0, 1},
	}.Mul(Mat4{
		{1, 0, 0, -center.X},
		{0, 1, 0, -center.Y},
		{0, 0, 1, -center.Z},
		{0, 0, 0, 1},
	})
}

// SetPerspective — центральная проекция с камерой на расстоянии f от плоскости z=0.
func (r *Renderer) SetPerspective(f float64) {
	r.Perspective = Mat4{
		{1, 0, 0, 0},
		{0, 1, 0, 0},
		{0, 0, 1, 0},
		{0, 0, -1 / f, 1},
	}
}

// SetViewport отображает куб [-1,1]³ в прямоугольник экрана [x,x+w]×[y,y+h].
func (r *Renderer) SetViewport(x, y, w, h int) {
	r.Viewport = Mat4{
		{float64(w) / 2, 0, 0, float64(x) + float64(w)/2},
		{0, float64(h) / 2, 0, float64(y) + float64(h)/2},
		{0, 0, 1, 0},
		{0, 0, 0, 1},
	}
}

// Rasterize закрашивает треугольник, заданный вершинами в clip-координатах.
func (r *Renderer) Rasterize(clip [3]Vec4, shader IShader) {
	framebuffer, zbuffer := r.Framebuffer, r.zbuffer
	var ndc [3]Vec4    // normalized device coordinates
	var screen [3]Vec2 // экранные координаты
	for i := range clip {
		ndc[i] = clip[i].Scale(1 / clip[i].W)
		screen[i] = r.Viewport.MulVec(ndc[i]).XY()
	}

	abc := Mat3{
		{screen[0].X, screen[0].Y, 1},
		{screen[1].X, screen[1].Y, 1},
		{screen[2].X, screen[2].Y, 1},
	}
	if abc.Det() < 1 {
		return // back-face culling + треугольники меньше пикселя
	}
	// Барицентрические координаты линейны по (x, y, 1): bc = (ABC⁻¹)ᵀ · (x, y, 1).
	// Матрицу считаем один раз на треугольник, а не на каждый пиксель.
	bcMatrix := abc.InvertTranspose()
	depth := Vec3{ndc[0].Z, ndc[1].Z, ndc[2].Z}

	// Bounding box, обрезанный по экрану.
	width, height := framebuffer.Width(), framebuffer.Height()
	minX := max(int(math.Min(screen[0].X, math.Min(screen[1].X, screen[2].X))), 0)
	maxX := min(int(math.Max(screen[0].X, math.Max(screen[1].X, screen[2].X))), width-1)
	minY := max(int(math.Min(screen[0].Y, math.Min(screen[1].Y, screen[2].Y))), 0)
	maxY := min(int(math.Max(screen[0].Y, math.Max(screen[1].Y, screen[2].Y))), height-1)
	if minX > maxX || minY > maxY {
		return
	}

	// Каждая горутина обрабатывает свой непрерывный диапазон строк, так что
	// пиксели framebuffer и ячейки zbuffer у горутин не пересекаются.
	rows := maxY - minY + 1
	workers := min(runtime.NumCPU(), rows)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		yStart := minY + w*rows/workers
		yEnd := minY + (w+1)*rows/workers // не включительно
		wg.Go(func() {
			for y := yStart; y < yEnd; y++ {
				for x := minX; x <= maxX; x++ {
					bc := bcMatrix.MulVec(Vec3{float64(x), float64(y), 1})
					if bc.X < 0 || bc.Y < 0 || bc.Z < 0 {
						continue // пиксель вне треугольника
					}
					z := bc.Dot(depth) // линейная интерполяция глубины
					idx := x + y*width
					if z <= zbuffer[idx] {
						continue // фрагмент глубже уже нарисованного
					}
					discard, color := shader.Fragment(bc)
					if discard {
						continue
					}
					zbuffer[idx] = z
					framebuffer.Set(x, y, color)
				}
			}
		})
	}
	wg.Wait()
}
